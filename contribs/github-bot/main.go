package main

import (
	"bot/client"
	"bot/param"
	"sync"

	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

func main() {
	// Retrieve params by parsing CLI flags and/or GitHub Actions context
	params := param.Get()

	// Init GitHub API client
	gh := client.New(params)

	// Handle comment update, if any
	handleCommentUpdate(gh)

	// Retrieve a slice of pull requests to process
	var (
		prs []*github.PullRequest
		err error
	)

	// If requested, retrieve all opened pull requests
	if params.PrAll {
		opts := &github.PullRequestListOptions{
			State:     "open",
			Sort:      "updated",
			Direction: "desc",
		}

		prs, _, err = gh.Client.PullRequests.List(gh.Ctx, gh.Owner, gh.Repo, opts)
		if err != nil {
			gh.Logger.Fatalf("Unable to retrieve all opened pull requests : %v", err)
		}

		// Otherwise, retrieve only specified pull request(s) (flag or GitHub Action context)
	} else {
		prs = make([]*github.PullRequest, len(params.PrNums))
		for i, prNum := range params.PrNums {
			pr, _, err := gh.Client.PullRequests.Get(gh.Ctx, gh.Owner, gh.Repo, prNum)
			if err != nil {
				gh.Logger.Fatalf("Unable to retrieve specified pull request (%d) : %v", prNum, err)
			}
			prs[i] = pr
		}
	}

	// Process all pull requests in parrallel
	autoRules, manualRules := config(gh)
	var wg sync.WaitGroup
	wg.Add(len(prs))

	for _, pr := range prs {
		go func(pr *github.PullRequest) {
			defer wg.Done()
			commentContent := CommentContent{}

			// Iterate over all automatic rules in config
			for _, autoRule := range autoRules {
				ifDetails := treeprint.NewWithRoot("🟢 Condition met")

				// Check if conditions of this rule are met by this PR
				if autoRule.If.IsMet(pr, ifDetails) {
					c := AutoContent{Description: autoRule.Description, Satisfied: false}
					thenDetails := treeprint.NewWithRoot("🔴 Requirement not satisfied")

					// Check if requirements of this rule are satisfied by this PR
					if autoRule.Then.IsSatisfied(pr, thenDetails) {
						thenDetails.SetValue("🟢 Requirement satisfied")
						c.Satisfied = true
					}

					c.ConditionDetails = ifDetails.String()
					c.RequirementDetails = thenDetails.String()
					commentContent.AutoRules = append(commentContent.AutoRules, c)
				}
			}

			// Iterate over all manual rules in config
			for _, manualRule := range manualRules {
				ifDetails := treeprint.NewWithRoot("🟢 Condition met")

				// Retrieve manual check states
				checks := make(map[string][2]string)
				if comment := gh.GetBotComment(pr.GetNumber()); comment != nil {
					checks = getCommentManualChecks(comment.GetBody())
				}

				// Check if conditions of this rule are met by this PR
				if manualRule.If.IsMet(pr, ifDetails) {
					commentContent.ManualRules = append(
						commentContent.ManualRules,
						ManualContent{
							Description:      manualRule.Description,
							ConditionDetails: ifDetails.String(),
							CheckedBy:        checks[manualRule.Description][1],
							Teams:            manualRule.Teams,
						},
					)
				}
			}

			// Print results in PR comment or in logs
			if gh.DryRun {
				// TODO: Pretty print dry run
			} else {
				updateComment(gh, pr, commentContent)
			}
		}(pr)
	}
	wg.Wait()
}
