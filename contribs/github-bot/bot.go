package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/logger"
	p "github.com/gnolang/gno/contribs/github-bot/params"
	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

func execBot(params *p.Params) error {
	// Create context with timeout if specified in the parameters
	ctx := context.Background()
	if params.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), params.Timeout)
		defer cancel()
	}

	// Init GitHub API client
	gh, err := client.New(ctx, params)
	if err != nil {
		return fmt.Errorf("comment update handling failed: %w", err)
	}

	// Handle comment update, if any
	if err := handleCommentUpdate(gh); err == errTriggeredByBot {
		return nil // Ignore if this run was triggered by a previous run
	} else if err != nil {
		return fmt.Errorf("comment update handling failed: %w", err)
	}

	// Retrieve a slice of pull requests to process
	var prs []*github.PullRequest

	// If requested, retrieve all open pull requests
	if params.PrAll {
		opts := &github.PullRequestListOptions{
			State:     "open",
			Sort:      "updated",
			Direction: "desc",
		}

		prs, _, err = gh.Client.PullRequests.List(gh.Ctx, gh.Owner, gh.Repo, opts)
		if err != nil {
			return fmt.Errorf("unable to retrieve all open pull requests: %w", err)
		}

		// Otherwise, retrieve only specified pull request(s)
		// (flag or GitHub Action context)
	} else {
		prs = make([]*github.PullRequest, len(params.PrNums))
		for i, prNum := range params.PrNums {
			pr, _, err := gh.Client.PullRequests.Get(gh.Ctx, gh.Owner, gh.Repo, prNum)
			if err != nil {
				return fmt.Errorf("unable to retrieve specified pull request (%d): %w", prNum, err)
			}
			prs[i] = pr
		}
	}

	if len(prs) > 1 {
		prNums := make([]int, len(prs))
		for i, pr := range prs {
			prNums[i] = pr.GetNumber()
		}

		gh.Logger.Infof("%d pull requests to process: %v\n", len(prNums), prNums)
	}

	// Process all pull requests in parallel
	autoRules, manualRules := config(gh)
	var wg sync.WaitGroup

	// Used in dry-run mode to log cleanly from different goroutines
	logMutex := sync.Mutex{}

	for _, pr := range prs {
		wg.Add(1)
		go func(pr *github.PullRequest) {
			defer wg.Done()
			commentContent := CommentContent{}
			commentContent.allSatisfied = true

			// Iterate over all automatic rules in config
			for _, autoRule := range autoRules {
				ifDetails := treeprint.NewWithRoot(fmt.Sprintf("%s Condition met", utils.StatusSuccess))

				// Check if conditions of this rule are met by this PR
				if !autoRule.If.IsMet(pr, ifDetails) {
					continue
				}

				c := AutoContent{Description: autoRule.Description, Satisfied: false}
				thenDetails := treeprint.NewWithRoot(fmt.Sprintf("%s Requirement not satisfied", utils.StatusFail))

				// Check if requirements of this rule are satisfied by this PR
				if autoRule.Then.IsSatisfied(pr, thenDetails) {
					thenDetails.SetValue(fmt.Sprintf("%s Requirement satisfied", utils.StatusSuccess))
					c.Satisfied = true
				} else {
					commentContent.allSatisfied = false
				}

				c.ConditionDetails = ifDetails.String()
				c.RequirementDetails = thenDetails.String()
				commentContent.AutoRules = append(commentContent.AutoRules, c)
			}

			// Iterate over all manual rules in config
			for _, manualRule := range manualRules {
				ifDetails := treeprint.NewWithRoot(fmt.Sprintf("%s Condition met", utils.StatusSuccess))

				// Retrieve manual check states
				checks := make(map[string][2]string)
				if comment := gh.GetBotComment(pr.GetNumber()); comment != nil {
					checks = getCommentManualChecks(comment.GetBody())
				}

				// Check if conditions of this rule are met by this PR
				if !manualRule.If.IsMet(pr, ifDetails) {
					continue
				}

				commentContent.ManualRules = append(
					commentContent.ManualRules,
					ManualContent{
						Description:      manualRule.Description,
						ConditionDetails: ifDetails.String(),
						CheckedBy:        checks[manualRule.Description][1],
						Teams:            manualRule.Teams,
					},
				)

				if checks[manualRule.Description][1] == "" {
					commentContent.allSatisfied = false
				}
			}

			// Logs results or write them in bot PR comment
			if gh.DryRun {
				logMutex.Lock()
				logResults(gh.Logger, pr.GetNumber(), commentContent)
				logMutex.Unlock()
			} else {
				updatePullRequest(gh, pr, commentContent)
			}
		}(pr)
	}
	wg.Wait()

	return nil
}

// logResults is called in dry-run mode and outputs the status of each check
// and a conclusion
func logResults(logger logger.Logger, prNum int, commentContent CommentContent) {
	logger.Infof("Pull request #%d requirements", prNum)
	if len(commentContent.AutoRules) > 0 {
		logger.Infof("Automated Checks:")
	}

	for _, rule := range commentContent.AutoRules {
		status := utils.StatusFail
		if rule.Satisfied {
			status = utils.StatusSuccess
		}
		logger.Infof("%s %s", status, rule.Description)
		logger.Debugf("If:\n%s", rule.ConditionDetails)
		logger.Debugf("Then:\n%s", rule.RequirementDetails)
	}

	if len(commentContent.ManualRules) > 0 {
		logger.Infof("Manual Checks:")
	}

	for _, rule := range commentContent.ManualRules {
		status := utils.StatusFail
		checker := "any user with comment edit permission"
		if rule.CheckedBy != "" {
			status = utils.StatusSuccess
		}
		if len(rule.Teams) == 0 {
			checker = fmt.Sprintf("a member of one of these teams: %s", strings.Join(rule.Teams, ", "))
		}
		logger.Infof("%s %s", status, rule.Description)
		logger.Debugf("If:\n%s", rule.ConditionDetails)
		logger.Debugf("Can be checked by %s", checker)
	}

	logger.Infof("Conclusion:")
	if commentContent.allSatisfied {
		logger.Infof("%s All requirements are satisfied\n", utils.StatusSuccess)
	} else {
		logger.Infof("%s Not all requirements are satisfied\n", utils.StatusFail)
	}
}
