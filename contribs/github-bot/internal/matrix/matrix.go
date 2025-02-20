package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/sethvargo/go-githubactions"
)

func execMatrix(flags *matrixFlags) error {
	// Get GitHub Actions context to retrieve event.
	actionCtx, err := githubactions.Context()
	if err != nil {
		return fmt.Errorf("unable to get GitHub Actions context: %w", err)
	}

	// If verbose is set, print the Github Actions event for debugging purpose.
	if *flags.verbose {
		jsonBytes, err := json.MarshalIndent(actionCtx.Event, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal event to json: %w", err)
		}
		fmt.Println("Event:", string(jsonBytes))
	}

	// Init Github client using only GitHub Actions context.
	owner, repo := actionCtx.Repo()
	gh, err := client.New(context.Background(), &client.Config{
		Owner:   owner,
		Repo:    repo,
		Verbose: *flags.verbose,
		DryRun:  true,
	})
	if err != nil {
		return fmt.Errorf("unable to init GitHub client: %w", err)
	}

	// Retrieve PR list from GitHub Actions event.
	prList, err := getPRListFromEvent(gh, actionCtx)
	if err != nil {
		return err
	}

	// Format PR list for GitHub Actions matrix definition.
	bytes, err := prList.MarshalText()
	if err != nil {
		return fmt.Errorf("unable to marshal PR list: %w", err)
	}
	matrix := fmt.Sprintf("%s=[%s]", flags.matrixKey, string(bytes))

	// If verbose is set, print the matrix for debugging purpose.
	if *flags.verbose {
		fmt.Printf("Matrix: %s\n", matrix)
	}

	// Get the path of the GitHub Actions environment file used for output.
	output, ok := os.LookupEnv("GITHUB_OUTPUT")
	if !ok {
		return errors.New("unable to get GITHUB_OUTPUT var")
	}

	// Open GitHub Actions output file
	file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open GitHub Actions output file: %w", err)
	}
	defer file.Close()

	// Append matrix to GitHub Actions output file
	if _, err := fmt.Fprintf(file, "%s\n", matrix); err != nil {
		return fmt.Errorf("unable to write matrix in GitHub Actions output file: %w", err)
	}

	return nil
}

func getPRListFromEvent(gh *client.GitHub, actionCtx *githubactions.GitHubContext) (utils.PRList, error) {
	var prList utils.PRList

	switch actionCtx.EventName {
	// Event triggered from GitHub Actions user interface.
	case utils.EventWorkflowDispatch:
		// Get input entered by the user.
		rawInput, ok := utils.IndexMap(actionCtx.Event, "inputs", "pull-request-list").(string)
		if !ok {
			return nil, errors.New("unable to get workflow dispatch input")
		}
		input := strings.TrimSpace(rawInput)

		// If all PR are requested, list them from GitHub API.
		if input == "all" {
			prs, err := gh.ListPR(utils.PRStateOpen)
			if err != nil {
				return nil, fmt.Errorf("unable to list all PR: %w", err)
			}

			prList = make(utils.PRList, len(prs))
			for i := range prs {
				prList[i] = prs[i].GetNumber()
			}
		} else {
			// If a PR list is provided, parse it.
			if err := prList.UnmarshalText([]byte(input)); err != nil {
				return nil, fmt.Errorf("invalid PR list provided as input: %w", err)
			}
		}

	// Event triggered by an issue / PR comment being created / edited / deleted
	// or any update on a PR.
	case utils.EventIssueComment, utils.EventPullRequest, utils.EventPullRequestReview, utils.EventPullRequestTarget:
		// For these events, retrieve the number of the associated PR from the context.
		prNum, err := utils.GetPRNumFromActionsCtx(actionCtx)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve PR number from GitHub Actions context: %w", err)
		}
		prList = utils.PRList{prNum}

	default:
		return nil, fmt.Errorf("unsupported event type: %s", actionCtx.EventName)
	}

	// Then only keep provided PR that are opened.
	var openedPRList utils.PRList = nil
	for _, prNum := range prList {
		if _, err := gh.GetOpenedPullRequest(prNum); err != nil {
			gh.Logger.Warningf("Can't get PR from event: %v", err)
		} else {
			openedPRList = append(openedPRList, prNum)
		}
	}

	return openedPRList, nil
}
