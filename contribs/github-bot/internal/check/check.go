package check

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/config"
	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
	"github.com/sethvargo/go-githubactions"
	"github.com/xlab/treeprint"
)

func execCheck(flags *checkFlags) error {
	// Create context with timeout if specified in the parameters.
	ctx := context.Background()
	if flags.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), flags.Timeout)
		defer cancel()
	}

	// Init GitHub API client.
	gh, err := client.New(ctx, &client.Config{
		Owner:   flags.Owner,
		Repo:    flags.Repo,
		Verbose: *flags.Verbose,
		DryRun:  flags.DryRun,
	})
	if err != nil {
		return fmt.Errorf("comment update handling failed: %w", err)
	}

	// Get GitHub Actions context to retrieve comment update.
	actionCtx, err := githubactions.Context()
	if err != nil {
		gh.Logger.Debugf("Unable to retrieve GitHub Actions context: %v", err)
		return nil
	}

	// Handle comment update, if any.
	if err := handleCommentUpdate(gh, actionCtx); errors.Is(err, errTriggeredByBot) {
		return nil // Ignore if this run was triggered by a previous run.
	} else if err != nil {
		return fmt.Errorf("comment update handling failed: %w", err)
	}

	// Retrieve a slice of pull requests to process.
	var prs []*github.PullRequest

	// If requested, retrieve all open pull requests.
	if flags.PRAll {
		prs, err = gh.ListPR(utils.PRStateOpen)
		if err != nil {
			return fmt.Errorf("unable to list all PR: %w", err)
		}
	} else {
		// Otherwise, retrieve only specified pull request(s)
		// (flag or GitHub Action context).
		prs = make([]*github.PullRequest, len(flags.PRNums))
		for i, prNum := range flags.PRNums {
			pr, err := gh.GetOpenedPullRequest(prNum)
			if err != nil {
				return fmt.Errorf("unable to process PR list: %w", err)
			}
			prs[i] = pr
		}
	}

	return processPRList(gh, prs)
}

func processPRList(gh *client.GitHub, prs []*github.PullRequest) error {
	if len(prs) > 1 {
		prNums := make([]int, len(prs))
		for i, pr := range prs {
			prNums[i] = pr.GetNumber()
		}

		gh.Logger.Infof("%d pull requests to process: %v\n", len(prNums), prNums)
	}

	// Process all pull requests in parallel.
	autoRules, manualRules := config.Config(gh)
	var wg sync.WaitGroup

	// Used in dry-run mode to log cleanly from different goroutines.
	logMutex := sync.Mutex{}

	// Used in regular-run mode to return an error if one PR processing failed.
	var failed atomic.Bool

	for _, pr := range prs {
		wg.Add(1)
		go func(pr *github.PullRequest) {
			defer wg.Done()
			commentContent := CommentContent{}
			commentContent.AutoAllSatisfied = true
			commentContent.ManualAllSatisfied = true

			// Iterate over all automatic rules in config.
			for _, autoRule := range autoRules {
				ifDetails := treeprint.NewWithRoot(fmt.Sprintf("%s Condition met", utils.Success))

				// Check if conditions of this rule are met by this PR.
				if !autoRule.If.IsMet(pr, ifDetails) {
					continue
				}

				c := AutoContent{Description: autoRule.Description, Satisfied: false}
				thenDetails := treeprint.NewWithRoot(fmt.Sprintf("%s Requirement not satisfied", utils.Fail))

				// Check if requirements of this rule are satisfied by this PR.
				if autoRule.Then.IsSatisfied(pr, thenDetails) {
					thenDetails.SetValue(fmt.Sprintf("%s Requirement satisfied", utils.Success))
					c.Satisfied = true
				} else {
					commentContent.AutoAllSatisfied = false
				}

				c.ConditionDetails = ifDetails.String()
				c.RequirementDetails = thenDetails.String()
				commentContent.AutoRules = append(commentContent.AutoRules, c)
			}

			// Retrieve manual check states.
			checks := make(map[string]manualCheckDetails)
			if comment, err := gh.GetBotComment(pr.GetNumber()); err == nil {
				checks = getCommentManualChecks(comment.GetBody())
			}

			// Iterate over all manual rules in config.
			for _, manualRule := range manualRules {
				ifDetails := treeprint.NewWithRoot(fmt.Sprintf("%s Condition met", utils.Success))

				// Check if conditions of this rule are met by this PR.
				if !manualRule.If.IsMet(pr, ifDetails) {
					continue
				}

				// Get check status from current comment, if any.
				checkedBy := ""
				check, ok := checks[manualRule.Description]
				if ok {
					checkedBy = check.checkedBy
				}

				commentContent.ManualRules = append(
					commentContent.ManualRules,
					ManualContent{
						Description:      manualRule.Description,
						ConditionDetails: ifDetails.String(),
						CheckedBy:        checkedBy,
						Teams:            manualRule.Teams,
					},
				)

				// If this check is the special one, store its state in the dedicated var.
				if manualRule.Description == config.ForceSkipDescription {
					if checkedBy != "" {
						commentContent.ForceSkip = true
					}
				} else if checkedBy == "" {
					// Or if its a normal check, just verify if it was checked by someone.
					commentContent.ManualAllSatisfied = false
				}
			}

			// Logs results or write them in bot PR comment.
			if gh.DryRun {
				logMutex.Lock()
				logResults(gh.Logger, pr.GetNumber(), commentContent)
				logMutex.Unlock()
			} else {
				if err := updatePullRequest(gh, pr, commentContent); err != nil {
					gh.Logger.Errorf("unable to update pull request: %v", err)
					failed.Store(true)
				}
			}
		}(pr)
	}
	wg.Wait()

	if failed.Load() {
		return errors.New("error occurred while processing pull requests")
	}

	return nil
}

// logResults is called in dry-run mode and outputs the status of each check
// and a conclusion.
func logResults(logger logger.Logger, prNum int, commentContent CommentContent) {
	logger.Infof("Pull request #%d requirements", prNum)
	if len(commentContent.AutoRules) > 0 {
		logger.Infof("Automated Checks:")
	}

	for _, rule := range commentContent.AutoRules {
		status := utils.Fail
		if rule.Satisfied {
			status = utils.Success
		}
		logger.Infof("%s %s", status, rule.Description)
		logger.Debugf("If:\n%s", rule.ConditionDetails)
		logger.Debugf("Then:\n%s", rule.RequirementDetails)
	}

	if len(commentContent.ManualRules) > 0 {
		logger.Infof("Manual Checks:")
	}

	for _, rule := range commentContent.ManualRules {
		status := utils.Fail
		checker := "any user with comment edit permission"
		if rule.CheckedBy != "" {
			status = utils.Success
		}
		if len(rule.Teams) == 0 {
			checker = fmt.Sprintf("a member of one of these teams: %s", strings.Join(rule.Teams, ", "))
		}
		logger.Infof("%s %s", status, rule.Description)
		logger.Debugf("If:\n%s", rule.ConditionDetails)
		logger.Debugf("Can be checked by %s", checker)
	}

	logger.Infof("Conclusion:")

	if commentContent.AutoAllSatisfied {
		logger.Infof("%s All automated checks are satisfied", utils.Success)
	} else {
		logger.Infof("%s Some automated checks are not satisfied", utils.Fail)
	}

	if commentContent.ManualAllSatisfied {
		logger.Infof("%s All manual checks are satisfied\n", utils.Success)
	} else {
		logger.Infof("%s Some manual checks are not satisfied\n", utils.Fail)
	}

	if commentContent.ForceSkip {
		logger.Infof("%s Bot checks are force skipped\n", utils.Success)
	}
}
