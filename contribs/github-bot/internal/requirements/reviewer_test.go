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

func Test_deduplicateReviews(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []*github.PullRequestReview
		expected []*github.PullRequestReview
	}{
		{
			name: "three different authors",
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("APPROVED")},
				{User: &github.User{Login: github.String("user2")}, State: github.String("CHANGES_REQUESTED")},
				{User: &github.User{Login: github.String("user3")}, State: github.String("COMMENTED")},
			},
			expected: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("APPROVED")},
				{User: &github.User{Login: github.String("user2")}, State: github.String("CHANGES_REQUESTED")},
				{User: &github.User{Login: github.String("user3")}, State: github.String("COMMENTED")},
			},
		},
		{
			name: "single author - approval then comment",
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("APPROVED")},
				{User: &github.User{Login: github.String("user1")}, State: github.String("COMMENTED")},
			},
			expected: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("APPROVED")},
			},
		},
		{
			name: "single author - approval then changes requested",
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("APPROVED")},
				{User: &github.User{Login: github.String("user1")}, State: github.String("CHANGES_REQUESTED")},
			},
			expected: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("CHANGES_REQUESTED")},
			},
		},
		{
			name: "two authors - mixed reviews",
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("userA")}, State: github.String("APPROVED")},
				{User: &github.User{Login: github.String("userB")}, State: github.String("CHANGES_REQUESTED")},
				{User: &github.User{Login: github.String("userA")}, State: github.String("CHANGES_REQUESTED")},
				{User: &github.User{Login: github.String("userB")}, State: github.String("COMMENTED")},
			},
			expected: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("userA")}, State: github.String("CHANGES_REQUESTED")},
				{User: &github.User{Login: github.String("userB")}, State: github.String("CHANGES_REQUESTED")},
			},
		},
		{
			name: "two authors - approval/changes requested then dismissed",
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: github.String("user1")}, State: github.String("APPROVED")},
				{User: &github.User{Login: github.String("user1")}, State: github.String("DISMISSED")},
				{User: &github.User{Login: github.String("user2")}, State: github.String("CHANGES_REQUESTED")},
				{User: &github.User{Login: github.String("user2")}, State: github.String("DISMISSED")},
			},
			expected: []*github.PullRequestReview{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateReviews(tt.reviews)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReviewByUser(t *testing.T) {
	t.Parallel()

	var (
		user1    = &github.User{Login: github.String("user1")}
		user2    = &github.User{Login: github.String("user2")}
		user3    = &github.User{Login: github.String("user3")}
		user4    = &github.User{Login: github.String("user4")}
		prAuthor = &github.User{Login: github.String("prAuthor")}

		reviewers = github.Reviewers{
			Users: []*github.User{user1, user2, user3},
		}

		reviews = []*github.PullRequestReview{
			{User: user1, State: github.String("APPROVED")},
			// Should be ignored in favour of the following one
			{User: user2, State: github.String("APPROVED")},
			{User: user2, State: github.String("CHANGES_REQUESTED")},
		}
	)

	for _, testCase := range []struct {
		name        string
		user        string
		action      RequestAction
		isSatisfied bool
		isRequested bool
	}{
		{"reviewer approved with RequestApply", user1.GetLogin(), RequestApply, true, false},
		{"reviewer approved with RequestIgnore", user1.GetLogin(), RequestIgnore, true, false},
		{"reviewer approved with RequestRemove", user1.GetLogin(), RequestRemove, true, false},

		{"reviewer requested changes with RequestApply", user2.GetLogin(), RequestApply, false, false},
		{"reviewer requested changes with RequestIgnore", user2.GetLogin(), RequestIgnore, false, false},
		{"reviewer requested changes with RequestRemove", user2.GetLogin(), RequestIgnore, false, false},

		{"reviewer not reviewed with RequestApply", user3.GetLogin(), RequestApply, false, false},
		{"reviewer not reviewed with RequestIgnore", user3.GetLogin(), RequestIgnore, false, false},
		{"reviewer not reviewed with RequestRemove", user3.GetLogin(), RequestRemove, false, true},

		{"not a reviewer with RequestApply", user4.GetLogin(), RequestApply, false, true},
		{"not a reviewer with RequestIgnore", user4.GetLogin(), RequestIgnore, false, false},
		{"not a reviewer with RequestRemove", user4.GetLogin(), RequestRemove, false, false},

		{"reviewer is the PR author with RequestApply", prAuthor.GetLogin(), RequestApply, false, false},
		{"reviewer is the PR author with RequestIgnore", prAuthor.GetLogin(), RequestIgnore, false, false},
		{"reviewer is the PR author with RequestRemove", prAuthor.GetLogin(), RequestRemove, false, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			firstRequest := true
			requested := false

			// This is a mock HTTP client that simulates the behavior of the GitHub API.
			mockedHTTPClient := mock.NewMockedHTTPClient(
				// This handler simulates the request reviewers endpoint used to
				// get the reviewers list or request a review.
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/0/requested_reviewers",
						Method:  "GET",
					},
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						// The first request is just used to get the reviewers.
						if firstRequest {
							w.Write(mock.MustMarshal(reviewers))
							firstRequest = false
						} else {
							// A subsequent request indicates that a review was requested
							// or a request was removed.
							requested = true
						}
					}),
				),
				// This handler simulates the reviews endpoint.
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/0/reviews",
						Method:  "GET",
					},
					reviews,
				),
			)

			// Create a new GitHub client with the mocked HTTP client.
			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			// Create a new PullRequest object with an author
			pr := &github.PullRequest{User: prAuthor}

			// Run the requirement and check if it is satisfied and if a review was requested.
			details := treeprint.New()
			requirement := ReviewByUser(gh, testCase.user, testCase.action).WithDesiredState(utils.ReviewStateApproved)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
			assert.Equal(t, testCase.isRequested, requested, fmt.Sprintf("requirement should have requested a review: %t", testCase.isRequested))
		})
	}
}

func TestReviewByTeamMembers(t *testing.T) {
	t.Parallel()

	var (
		user1 = &github.User{Login: github.String("user1")}
		user2 = &github.User{Login: github.String("user2")}
		user3 = &github.User{Login: github.String("user3")}
		user4 = &github.User{Login: github.String("user4")}
		user5 = &github.User{Login: github.String("user5")}

		team1 = &github.Team{Slug: github.String("team1")}
		team2 = &github.Team{Slug: github.String("team2")}
		team3 = &github.Team{Slug: github.String("team3")}

		noReviewers   = github.Reviewers{}
		teamReviewers = github.Reviewers{
			Teams: []*github.Team{team1, team2, team3},
		}
		userReviewers = github.Reviewers{
			Users: []*github.User{user1, user2, user3, user4, user5},
		}

		members = map[string][]*github.User{
			team1.GetSlug(): {user1, user2, user3},
			team2.GetSlug(): {user3, user4, user5},
			team3.GetSlug(): {user4, user5},
		}

		reviews = []*github.PullRequestReview{
			// Only later review should be counted (user1).
			{User: user1, State: github.String("CHANGES_REQUESTED")},
			{User: user1, State: github.String("APPROVED")},
			{User: user2, State: github.String("APPROVED")},
			{User: user3, State: github.String("APPROVED")},
			{User: user4, State: github.String("CHANGES_REQUESTED")},
			// Only later review should be counted (user5).
			{User: user5, State: github.String("APPROVED")},
			{User: user5, State: github.String("CHANGES_REQUESTED")},
		}
	)

	const (
		notSatisfied = 0
		satisfied    = 1
		withRequest  = 2
	)

	for _, testCase := range []struct {
		name           string
		req            *ReviewByTeamMembersRequirement
		reviews        []*github.PullRequestReview
		reviewers      github.Reviewers
		expectedResult byte
	}{
		{
			name: "3/3 team members approved with RequestApply",
			req: ReviewByTeamMembers(nil, "team1", RequestApply).
				WithCount(3).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied,
		},
		{
			name: "3/3 team members approved with RequestIgnore",
			req: ReviewByTeamMembers(nil, "team1", RequestIgnore).
				WithCount(3).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied,
		},
		{
			name: "3/3 team members approved with RequestRemove",
			req: ReviewByTeamMembers(nil, "team1", RequestRemove).
				WithCount(3).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied | withRequest,
		},
		{
			name: "3/3 team members approved (with user reviewers)",
			req: ReviewByTeamMembers(nil, "team1", RequestApply).
				WithCount(3).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      userReviewers,
			expectedResult: satisfied,
		},
		{
			name: "1/1 team member approved with RequestApply",
			req: ReviewByTeamMembers(nil, "team2", RequestApply).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied,
		},
		{
			name: "1/1 team member approved with RequestIgnore",
			req: ReviewByTeamMembers(nil, "team2", RequestRemove).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied | withRequest,
		},
		{
			name: "1/2 team member approved with RequestApply",
			req: ReviewByTeamMembers(nil, "team2", RequestApply).
				WithCount(2).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied,
		},
		{
			name: "1/2 team member approved with RequestIgnore",
			req: ReviewByTeamMembers(nil, "team2", RequestIgnore).
				WithCount(2).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied,
		},
		{
			name: "1/2 team member approved with RequestRemove",
			req: ReviewByTeamMembers(nil, "team2", RequestRemove).
				WithCount(2).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied | withRequest,
		},
		{
			name: "0/1 team member approved with RequestApply",
			req: ReviewByTeamMembers(nil, "team3", RequestApply).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied,
		},
		{
			name: "0/1 team member approved with RequestIgnore",
			req: ReviewByTeamMembers(nil, "team3", RequestIgnore).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied,
		},
		{
			name: "0/1 team member approved with RequestRemove",
			req: ReviewByTeamMembers(nil, "team3", RequestRemove).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied | withRequest,
		},
		{
			name: "0/1 team member reviewed with request with RequestApply",
			req:  ReviewByTeamMembers(nil, "team3", RequestApply),
			// Show there are no current reviews, so we actually perform the request.
			reviewers:      noReviewers,
			expectedResult: notSatisfied | withRequest,
		},
		{
			name:           "0/1 team member reviewed with request with RequestIgnore",
			req:            ReviewByTeamMembers(nil, "team3", RequestIgnore),
			reviewers:      noReviewers,
			expectedResult: notSatisfied,
		},
		{
			name:           "0/1 team member reviewed with request with RequestRemove",
			req:            ReviewByTeamMembers(nil, "team3", RequestRemove),
			reviewers:      noReviewers,
			expectedResult: notSatisfied,
		},
		{
			name: "3/3 team member approved from review list",
			req: ReviewByTeamMembers(nil, "team1", RequestApply).
				WithDesiredState(utils.ReviewStateApproved).
				WithCount(3),
			reviews:        reviews,
			reviewers:      noReviewers,
			expectedResult: satisfied,
		},
		{
			name: "1/2 team member approved from review list",
			req: ReviewByTeamMembers(nil, "team3", RequestApply).
				WithDesiredState(utils.ReviewStateApproved).
				WithCount(2),
			reviews:        reviews,
			reviewers:      noReviewers,
			expectedResult: notSatisfied,
		},
		{
			name: "team doesn't exist with request",
			req: ReviewByTeamMembers(nil, "team4", RequestApply).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      noReviewers,
			expectedResult: notSatisfied | withRequest,
		},
		{
			name: "3/3 team members reviewed",
			req: ReviewByTeamMembers(nil, "team2", RequestApply).
				WithCount(3),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied,
		},
		{
			name: "2/2 team members rejected with RequestApply",
			req: ReviewByTeamMembers(nil, "team2", RequestApply).
				WithCount(2).
				WithDesiredState(utils.ReviewStateChangesRequested),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied,
		},
		{
			name: "2/2 team members rejected with RequestIgnore",
			req: ReviewByTeamMembers(nil, "team2", RequestIgnore).
				WithCount(2).
				WithDesiredState(utils.ReviewStateChangesRequested),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied,
		},
		{
			name: "2/2 team members rejected with RequestRemove",
			req: ReviewByTeamMembers(nil, "team2", RequestRemove).
				WithCount(2).
				WithDesiredState(utils.ReviewStateChangesRequested),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: satisfied | withRequest,
		},
		{
			name: "1/3 team members approved",
			req: ReviewByTeamMembers(nil, "team2", RequestApply).
				WithCount(3).
				WithDesiredState(utils.ReviewStateApproved),
			reviews:        reviews,
			reviewers:      teamReviewers,
			expectedResult: notSatisfied,
		},
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
						Pattern: fmt.Sprintf("/orgs/teams/%s/members", testCase.req.team),
						Method:  "GET",
					},
					members[testCase.req.team],
				),
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/0/reviews",
						Method:  "GET",
					},
					testCase.reviews,
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			pr := &github.PullRequest{}
			details := treeprint.New()
			req := new(ReviewByTeamMembersRequirement)
			*req = *testCase.req
			req.gh = gh

			expSatisfied := testCase.expectedResult&satisfied > 0
			expRequested := testCase.expectedResult&withRequest > 0
			assert.Equal(t, expSatisfied, req.IsSatisfied(pr, details),
				"requirement should have a satisfied status: %t", expSatisfied)
			assert.True(t, utils.TestLastNodeStatus(t, expSatisfied, details),
				"requirement details should have a status: %t", expSatisfied)
			assert.Equal(t, expRequested, requested,
				"requirement should have requested a review: %t", expRequested)
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
			// should be ignored in favour of the following one.
			User:              &github.User{Login: github.String("user2")},
			State:             github.String("CHANGES_REQUESTED"),
			AuthorAssociation: github.String("COLLABORATOR"),
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

func TestReviewByAnyUser(t *testing.T) {
	t.Parallel()

	reviews := []*github.PullRequestReview{
		{
			User:  &github.User{Login: github.String("user1")},
			State: github.String("APPROVED"),
		}, {
			// should be ignored in favour of the following one.
			User:  &github.User{Login: github.String("user2")},
			State: github.String("CHANGES_REQUESTED"),
		}, {
			User:  &github.User{Login: github.String("user2")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("user3")},
			State: github.String("APPROVED"),
		}, {
			User:  &github.User{Login: github.String("user4")},
			State: github.String("CHANGES_REQUESTED"),
		},
	}

	for _, testCase := range []struct {
		name         string
		users        []string
		desiredState utils.ReviewState
		isSatisfied  bool
	}{
		{"empty users", []string{}, utils.ReviewStateApproved, false},
		{"non-matching user approval", []string{"user4"}, utils.ReviewStateApproved, false},
		{"matching approval", []string{"user2", "user4"}, utils.ReviewStateApproved, true},
		{"matching non-approval", []string{"user4"}, "", true},
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
			requirement := ReviewByAnyUser(gh, testCase.users...).
				WithDesiredState(testCase.desiredState)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}
