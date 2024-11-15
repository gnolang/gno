package conditions

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
		name   string
		user   string
		author string
		isMet  bool
	}{
		{"author match", "user", "user", true},
		{"author doesn't match", "user", "author", false},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{
				User: &github.User{Login: github.String(testCase.author)},
			}
			details := treeprint.New()
			condition := Author(testCase.user)

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
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
		name    string
		user    string
		members []*github.User
		isMet   bool
	}{
		{"empty member list", "user", []*github.User{}, false},
		{"member list contains user", "user", members, true},
		{"member list doesn't contain user", "user2", members, false},
	} {
		testCase := testCase
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
			condition := AuthorInTeam(gh, "team")

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}
