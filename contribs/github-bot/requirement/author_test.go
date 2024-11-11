package requirement

import (
	"context"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/logger"
	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v64/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/xlab/treeprint"
)

func TestAuthor(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		user        string
		author      string
		isSatisfied bool
	}{
		{"author match", "user", "user", true},
		{"author doesn't match", "user", "author", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{
				User: &github.User{Login: github.String(testCase.author)},
			}
			details := treeprint.New()
			requirement := Author(testCase.user)

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
		})
	}
}

func TestAuthorInTeam(t *testing.T) {
	t.Parallel()

	members := []*github.User{
		{Login: github.String("notTheRightOne")},
		{Login: github.String("user")},
		{Login: github.String("anotherOne")},
	}

	for _, testCase := range []struct {
		name        string
		user        string
		members     []*github.User
		isSatisfied bool
	}{
		{"empty assignee list", "user", []*github.User{}, false},
		{"assignee list contains user", "user", members, true},
		{"assignee list doesn't contain user", "user2", members, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: "/orgs/teams/team/members",
						Method:  "GET",
					},
					testCase.members,
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			pr := &github.PullRequest{
				User: &github.User{Login: github.String(testCase.user)},
			}
			details := treeprint.New()
			requirement := AuthorInTeam(gh, "team")

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
		})
	}
}
