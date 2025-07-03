package conditions

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"

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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{Assignees: testCase.assignees}
			details := treeprint.New()
			condition := Assignee(testCase.user)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
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

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
		})
	}
}
