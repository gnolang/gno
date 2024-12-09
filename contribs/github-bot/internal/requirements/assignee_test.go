package requirements

import (
	"context"
	"net/http"
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
		dryRun    bool
		exists    bool
	}{
		{"empty assignee list", "user", []*github.User{}, false, false},
		{"empty assignee list with dry-run", "user", []*github.User{}, true, false},
		{"assignee list contains user", "user", assignees, false, true},
		{"assignee list doesn't contain user", "user2", assignees, false, false},
		{"assignee list doesn't contain user with dry-run", "user2", assignees, true, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			requested := false
			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/issues/0/assignees",
						Method:  "GET", // It looks like this mock package doesn't support mocking POST requests.
					},
					http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
						requested = true
					}),
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
				DryRun: testCase.dryRun,
			}

			pr := &github.PullRequest{Assignees: testCase.assignees}
			details := treeprint.New()
			requirement := Assignee(gh, testCase.user)

			assert.True(t, requirement.IsSatisfied(pr, details) || testCase.dryRun, "requirement should have a satisfied status: true")
			assert.True(t, utils.TestLastNodeStatus(t, true, details) || testCase.dryRun, "requirement details should have a status: true")
			assert.True(t, testCase.exists || requested || testCase.dryRun, "requirement should have requested to create item")
		})
	}
}
