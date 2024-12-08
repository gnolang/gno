package matrix

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestProcessEvent(t *testing.T) {
	t.Parallel()

	prs := []*github.PullRequest{
		{Number: github.Int(1), State: github.String(utils.PRStateOpen)},
		{Number: github.Int(2), State: github.String(utils.PRStateOpen)},
		{Number: github.Int(3), State: github.String(utils.PRStateOpen)},
		{Number: github.Int(4), State: github.String(utils.PRStateClosed)},
		{Number: github.Int(5), State: github.String(utils.PRStateClosed)},
		{Number: github.Int(6), State: github.String(utils.PRStateClosed)},
	}
	openPRs := prs[:3]

	for _, testCase := range []struct {
		name           string
		gaCtx          *githubactions.GitHubContext
		prs            []*github.PullRequest
		expectedPRList utils.PRList
		expectedError  bool
	}{
		{
			"valid issue_comment event",
			&githubactions.GitHubContext{
				EventName: utils.EventIssueComment,
				Event:     map[string]any{"issue": map[string]any{"number": 1.}},
			},
			prs,
			utils.PRList{1},
			false,
		}, {
			"valid pull_request event",
			&githubactions.GitHubContext{
				EventName: utils.EventPullRequest,
				Event:     map[string]any{"pull_request": map[string]any{"number": 1.}},
			},
			prs,
			utils.PRList{1},
			false,
		}, {
			"valid pull_request_review event",
			&githubactions.GitHubContext{
				EventName: utils.EventPullRequestReview,
				Event:     map[string]any{"pull_request": map[string]any{"number": 1.}},
			},
			prs,
			utils.PRList{1},
			false,
		}, {
			"valid pull_request_target event",
			&githubactions.GitHubContext{
				EventName: utils.EventPullRequestTarget,
				Event:     map[string]any{"pull_request": map[string]any{"number": 1.}},
			},
			prs,
			utils.PRList{1},
			false,
		}, {
			"invalid event (PR number not set)",
			&githubactions.GitHubContext{
				EventName: utils.EventIssueComment,
				Event:     map[string]any{"issue": nil},
			},
			prs,
			utils.PRList(nil),
			true,
		}, {
			"invalid event name",
			&githubactions.GitHubContext{
				EventName: "invalid_event",
				Event:     map[string]any{"issue": map[string]any{"number": 1.}},
			},
			prs,
			utils.PRList(nil),
			true,
		}, {
			"valid workflow_dispatch all",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "all"}},
			},
			openPRs,
			utils.PRList{1, 2, 3},
			false,
		}, {
			"valid workflow_dispatch all (no prs)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "all"}},
			},
			nil,
			utils.PRList(nil),
			false,
		}, {
			"valid workflow_dispatch list",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "1,2,3"}},
			},
			prs,
			utils.PRList{1, 2, 3},
			false,
		}, {
			"valid workflow_dispatch list with spaces",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "    1,  2     ,3 "}},
			},
			prs,
			utils.PRList{1, 2, 3},
			false,
		}, {
			"invalid workflow_dispatch list (1 closed)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "1,2,3,4"}},
			},
			prs,
			utils.PRList{1, 2, 3},
			false,
		}, {
			"invalid workflow_dispatch list (1 doesn't exist)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "42"}},
			},
			prs,
			utils.PRList(nil),
			false,
		}, {
			"invalid workflow_dispatch list (all closed)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "4,5,6"}},
			},
			prs,
			utils.PRList(nil),
			false,
		}, {
			"invalid workflow_dispatch list (empty)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": ""}},
			},
			prs,
			utils.PRList(nil),
			true,
		}, {
			"invalid workflow_dispatch list (unset)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": ""},
			},
			prs,
			utils.PRList(nil),
			true,
		}, {
			"invalid workflow_dispatch list (not a number list)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "foo"}},
			},
			prs,
			utils.PRList(nil),
			true,
		}, {
			"invalid workflow_dispatch list (number list with invalid elem)",
			&githubactions.GitHubContext{
				EventName: utils.EventWorkflowDispatch,
				Event:     map[string]any{"inputs": map[string]any{"pull-request-list": "1,2,foo"}},
			},
			prs,
			utils.PRList(nil),
			true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/pulls",
						Method:  "GET",
					},
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						if testCase.expectedPRList != nil {
							w.Write(mock.MustMarshal(testCase.prs))
						}
					}),
				),
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/{number}",
						Method:  "GET",
					},
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						var (
							err   error
							prNum int
							parts = strings.Split(req.RequestURI, "/")
						)

						if len(parts) > 0 {
							prNumStr := parts[len(parts)-1]
							prNum, err = strconv.Atoi(prNumStr)
							if err != nil {
								panic(err) // Should never happen.
							}
						}

						for _, pr := range prs {
							if pr.GetNumber() == prNum {
								w.Write(mock.MustMarshal(pr))
								return
							}
						}

						w.Write(mock.MustMarshal(nil))
					}),
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			prList, err := getPRListFromEvent(gh, testCase.gaCtx)
			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedPRList, prList)
		})
	}
}
