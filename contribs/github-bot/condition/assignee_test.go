package condition

import (
	"context"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/logger"
	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/migueleliasweb/go-github-mock/src/mock"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

func TestAssignee(t *testing.T) {
	t.Parallel()

	assignees := []*github.User{
		{Login: github.String("notTheRightOne")},
		{Login: github.String("user")},
		{Login: github.String("anotherOne")},
	}

	for _, testCase := range []struct {
		name      string
		user      string
		assignees []*github.User
		isMet     bool
	}{
		{"empty assignee list", "user", []*github.User{}, false},
		{"assignee list contains user", "user", assignees, true},
		{"assignee list doesn't contain user", "user2", assignees, false},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{Assignees: testCase.assignees}
			details := treeprint.New()
			condition := Assignee(testCase.user)

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}

func TestAssigneeInTeam(t *testing.T) {
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
		{"empty assignee list", "user", []*github.User{}, false},
		{"assignee list contains user", "user", members, true},
		{"assignee list doesn't contain user", "user2", members, false},
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
				Assignees: []*github.User{
					{Login: github.String(testCase.user)},
				},
			}
			details := treeprint.New()
			condition := AssigneeInTeam(gh, "team")

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}
