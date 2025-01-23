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
			State: github.String("CHANGES_REQUESTED"),
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
			requirement := ReviewByUser(gh, testCase.user).WithDesiredState(utils.ReviewStateApproved)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
			assert.Equal(t, testCase.create, requested, fmt.Sprintf("requirement should have requested to create item: %t", testCase.create))
		})
	}
}

func TestReviewByTeamMembers(t *testing.T) {
	t.Parallel()

	var (
		reviewers = github.Reviewers{
			Teams: []*github.Team{
				{Slug: github.String("team1")},
				{Slug: github.String("team2")},
				{Slug: github.String("team3")},
			},
		}
		noReviewers   = github.Reviewers{}
		userReviewers = github.Reviewers{
			Users: []*github.User{
				{Login: github.String("user1")},
				{Login: github.String("user2")},
				{Login: github.String("user3")},
				{Login: github.String("user4")},
				{Login: github.String("user5")},
			},
		}
	)

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
			State: github.String("CHANGES_REQUESTED"),
		}, {
			User:  &github.User{Login: github.String("user5")},
			State: github.String("CHANGES_REQUESTED"),
		},
	}

	const (
		notSatisfied = 0
		satisfied    = 1
		withRequest  = 2
	)

	for _, testCase := range []struct {
		name           string
		team           string
		count          uint
		state          utils.ReviewState
		reviewers      github.Reviewers
		expectedResult byte
	}{
		{"3/3 team members approved", "team1", 3, utils.ReviewStateApproved, reviewers, satisfied},
		{"3/3 team members approved (with user reviewers)", "team1", 3, utils.ReviewStateApproved, userReviewers, satisfied},
		{"1/1 team member approved", "team2", 1, utils.ReviewStateApproved, reviewers, satisfied},
		{"1/2 team member approved", "team2", 2, utils.ReviewStateApproved, reviewers, notSatisfied},
		{"0/1 team member approved", "team3", 1, utils.ReviewStateApproved, reviewers, notSatisfied},
		{"0/1 team member approved with request", "team3", 1, utils.ReviewStateApproved, noReviewers, notSatisfied | withRequest},
		{"team doesn't exist with request", "team4", 1, utils.ReviewStateApproved, noReviewers, notSatisfied | withRequest},
		{"3/3 team members reviewed", "team2", 3, "", reviewers, satisfied},
		{"2/2 team members rejected", "team2", 2, utils.ReviewStateChangesRequested, reviewers, satisfied},
		{"1/3 team members approved", "team2", 3, utils.ReviewStateApproved, reviewers, notSatisfied},
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
							w.Write(mock.MustMarshal(testCase.reviewers))
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
			requirement := ReviewByTeamMembers(gh, testCase.team).
				WithCount(testCase.count).
				WithDesiredState(testCase.state)

			expSatisfied := testCase.expectedResult&satisfied != 0
			expRequested := testCase.expectedResult&withRequest > 0
			assert.Equal(t, expSatisfied, requirement.IsSatisfied(pr, details),
				"requirement should have a satisfied status: %t", expSatisfied)
			assert.True(t, utils.TestLastNodeStatus(t, expSatisfied, details),
				"requirement details should have a status: %t", expSatisfied)
			assert.Equal(t, expRequested, requested,
				"requirement should have requested to create item: %t", expRequested)
		})
	}
}

func TestReviewByOrgMembers(t *testing.T) {
	t.Parallel()

	reviews := []*github.PullRequestReview{
		{
			User:              &github.User{Login: github.String("user1")},
			State:             github.String("APPROVED"),
			AuthorAssociation: github.String("MEMBER"),
		}, {
			User:              &github.User{Login: github.String("user2")},
			State:             github.String("APPROVED"),
			AuthorAssociation: github.String("COLLABORATOR"),
		}, {
			User:              &github.User{Login: github.String("user3")},
			State:             github.String("APPROVED"),
			AuthorAssociation: github.String("MEMBER"),
		}, {
			User:              &github.User{Login: github.String("user4")},
			State:             github.String("CHANGES_REQUESTED"),
			AuthorAssociation: github.String("MEMBER"),
		}, {
			User:              &github.User{Login: github.String("user5")},
			State:             github.String("CHANGES_REQUESTED"),
			AuthorAssociation: github.String("NONE"),
		},
	}

	for _, testCase := range []struct {
		name         string
		count        uint
		desiredState utils.ReviewState
		isSatisfied  bool
	}{
		{"2/3 org members approved", 3, utils.ReviewStateApproved, false},
		{"2/2 org members approved", 2, utils.ReviewStateApproved, true},
		{"2/1 org members approved", 1, utils.ReviewStateApproved, true},
		{"3/3 org members reviewed", 3, "", true},
		{"3/4 org members reviewed", 4, "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockedHTTPClient := mock.NewMockedHTTPClient(
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
			requirement := ReviewByOrgMembers(gh).
				WithCount(testCase.count).
				WithDesiredState(testCase.desiredState)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}
