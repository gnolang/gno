package conditions

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/stretchr/testify/assert"

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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{
				User: &github.User{Login: github.String(testCase.author)},
			}
			details := treeprint.New()
			condition := Author(testCase.user)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
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

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
		})
	}
}

func TestAuthorAssociationIs(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name            string
		association     string
		associationWant string
		isMet           bool
	}{
		{"has", "MEMBER", "MEMBER", true},
		{"hasNot", "COLLABORATOR", "MEMBER", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{
				AuthorAssociation: github.String(testCase.association),
			}
			details := treeprint.New()
			condition := AuthorAssociationIs(testCase.associationWant)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
		})
	}
}
