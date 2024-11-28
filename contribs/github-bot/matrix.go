package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/params"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/sethvargo/go-githubactions"
)

func newMatrixCmd() *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "matrix",
			ShortUsage: "github-bot matrix",
			ShortHelp:  "parses GitHub Actions event and defines matrix accordingly",
			LongHelp:   "This tool checks if the requirements for a PR to be merged are satisfied (defined in config.go) and displays PR status checks accordingly.\nA valid GitHub Token must be provided by setting the GITHUB_TOKEN environment variable.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, _ []string) error {
			return execMatrix()
		},
	)
}

func execMatrix() error {
	// Get GitHub Actions context to retrieve event.
	actionCtx, err := githubactions.Context()
	if err != nil {
		return fmt.Errorf("unable to get GitHub Actions context: %w", err)
	}

	// Init Github client using only GitHub Actions context
	owner, repo := actionCtx.Repo()
	gh, err := client.New(context.Background(), &params.Params{Owner: owner, Repo: repo})
	if err != nil {
		return fmt.Errorf("unable to init GitHub client: %w", err)
	}

	// Retrieve PR list from GitHub Actions event
	prList, err := getPRListFromEvent(gh, actionCtx)
	if err != nil {
		return err
	}

	// Print PR list for GitHub Actions matrix definition
	bytes, err := prList.MarshalText()
	if err != nil {
		return fmt.Errorf("unable to marshal PR list: %w", err)
	}
	fmt.Printf("[%s]", string(bytes))

	return nil
}

func getPRListFromEvent(gh *client.GitHub, actionCtx *githubactions.GitHubContext) (params.PRList, error) {
	var prList params.PRList

	switch actionCtx.EventName {
	// Event triggered from GitHub Actions user interface
	case utils.EventWorkflowDispatch:
		// Get input entered by the user
		rawInput, ok := utils.IndexMap(actionCtx.Event, "inputs", "pull-request-list").(string)
		if !ok {
			return nil, errors.New("unable to get workflow dispatch input")
		}
		input := strings.TrimSpace(rawInput)

		// If all PR are requested, list them from GitHub API
		if input == "all" {
			prs, err := gh.ListPR(utils.PRStateOpen)
			if err != nil {
				return nil, fmt.Errorf("unable to list all PR: %w", err)
			}

			prList = make(params.PRList, len(prs))
			for i := range prs {
				prList[i] = prs[i].GetNumber()
			}
		} else {
			// If a PR list is provided, parse it
			if err := prList.UnmarshalText([]byte(input)); err != nil {
				return nil, fmt.Errorf("invalid PR list provided as input: %w", err)
			}

			// Then check if all provided PR are opened
			for _, prNum := range prList {
				pr, _, err := gh.Client.PullRequests.Get(gh.Ctx, gh.Owner, gh.Repo, prNum)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve specified pull request (%d): %w", prNum, err)
				} else if pr.GetState() != utils.PRStateOpen {
					return nil, fmt.Errorf("pull request %d is not opened, actual state: %s", prNum, pr.GetState())
				}
			}
		}

	// Event triggered by an issue / PR comment being created / edited / deleted
	// or any update on a PR
	case utils.EventIssueComment, utils.EventPullRequest, utils.EventPullRequestTarget:
		// For these events, retrieve the number of the associated PR from the context
		prNum, err := utils.GetPRNumFromActionsCtx(actionCtx)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve PR number from GitHub Actions context: %w", err)
		}
		prList = params.PRList{prNum}

	default:
		return nil, fmt.Errorf("unsupported event type: %s", actionCtx.EventName)
	}

	return prList, nil
}
