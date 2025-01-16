package requirements

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v64/github"
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
		{"reviewer matches", "user", true, false},
		{"reviewer matches without approval", "anotherOne", false, false},
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

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
			assert.Equal(t, testCase.create, requested, fmt.Sprintf("requirement should have requested to create item: %t", testCase.create))
		})
	}
}

func TestReviewByTeamMembers(t *testing.T) {
	t.Parallel()

	reviewers := github.Reviewers{
		Teams: []*github.Team{
			{Slug: github.String("team1")},
			{Slug: github.String("team2")},
			{Slug: github.String("team3")},
		},
	}

	members := map[string][]*github.User{
		"team1": {
			{Login: github.String("user1")},
			{Login: github.String("user2")},
			{Login: github.String("user3")},
		},
		"team2": {
			{Login: github.String("user3")},
			{Login: github.String("user4")},
			{Login: github.String("user5")},
		},
		"team3": {
			{Login: github.String("user4")},
			{Login: github.String("user5")},
		},
	}

	reviews := []*github.PullRequestReview{
		{
			User:  &github.User{Login: github.String("user1")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("user2")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("user3")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("user4")},
			State: github.String("REQUEST_CHANGES"),
		}, {
			User:  &github.User{Login: github.String("user5")},
			State: github.String("REQUEST_CHANGES"),
		},
	}

	for _, testCase := range []struct {
		name        string
		team        string
		count       uint
		isSatisfied bool
		testRequest bool
	}{
		{"3/3 team members approved;", "team1", 3, true, false},
		{"1/1 team member approved", "team2", 1, true, false},
		{"1/2 team member approved", "team2", 2, false, false},
		{"0/1 team member approved", "team3", 1, false, false},
		{"0/1 team member approved with request", "team3", 1, false, true},
		{"team doesn't exist with request", "team4", 1, false, true},
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
							if testCase.testRequest {
								w.Write(mock.MustMarshal(github.Reviewers{}))
							} else {
								w.Write(mock.MustMarshal(reviewers))
							}
							firstRequest = false
						} else {
							requested = true
						}
					}),
				),
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: fmt.Sprintf("/orgs/teams/%s/members", testCase.team),
						Method:  "GET",
					},
					members[testCase.team],
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
			requirement := ReviewByTeamMembers(gh, testCase.team, testCase.count)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
			assert.Equal(t, testCase.testRequest, requested, fmt.Sprintf("requirement should have requested to create item: %t", testCase.testRequest))
		})
	}
}
