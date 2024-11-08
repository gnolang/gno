package requirement

import (
	"context"
	"net/http"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/logger"
	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v66/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/xlab/treeprint"
)

func TestReviewByUser(t *testing.T) {
	t.Parallel()

	reviewers := github.Reviewers{
		Users: []*github.User{
			{Login: github.String("notTheRightOne")},
			{Login: github.String("user")},
			{Login: github.String("anotherOne")},
		},
	}

	reviews := []*github.PullRequestReview{
		{
			User:  &github.User{Login: github.String("notTheRightOne")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("user")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("anotherOne")},
			State: github.String("REQUEST_CHANGES"),
		},
	}

	for _, testCase := range []struct {
		name        string
		user        string
		isSatisfied bool
		create      bool
	}{
		{"reviewer match", "user", true, false},
		{"reviewer match without approval", "anotherOne", false, false},
		{"reviewer doesn't match", "user2", false, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			firstRequest := true
			requested := false
			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/0/requested_reviewers",
						Method:  "GET",
					},
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						if firstRequest {
							w.Write(mock.MustMarshal(reviewers))
							firstRequest = false
						} else {
							requested = true
						}
					}),
				),
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/0/reviews",
						Method:  "GET",
					},
					reviews,
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := ReviewByUser(gh, testCase.user)

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
			if testCase.create != requested {
				t.Errorf("requirement should have requested to create item: %t", testCase.create)
			}
		})
	}
}

func TestReviewByTeamMembers(t *testing.T) {
	t.Parallel()

	// TODO
}
