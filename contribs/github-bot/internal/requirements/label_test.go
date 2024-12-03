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

func TestLabel(t *testing.T) {
	t.Parallel()

	labels := []*github.Label{
		{Name: github.String("notTheRightOne")},
		{Name: github.String("label")},
		{Name: github.String("anotherOne")},
	}

	for _, testCase := range []struct {
		name    string
		pattern string
		labels  []*github.Label
		dryRun  bool
		exists  bool
	}{
		{"empty label list", "label", []*github.Label{}, false, false},
		{"empty label list with dry-run", "user", []*github.Label{}, true, false},
		{"label list contains label", "label", labels, false, true},
		{"label list doesn't contain label", "label2", labels, false, false},
		{"label list doesn't contain label with dry-run", "label", labels, true, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			requested := false
			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/issues/0/labels",
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

			pr := &github.PullRequest{Labels: testCase.labels}
			details := treeprint.New()
			requirement := Label(gh, testCase.pattern)

			assert.True(t, requirement.IsSatisfied(pr, details) || testCase.dryRun, "requirement should have a satisfied status: true")
			assert.True(t, utils.TestLastNodeStatus(t, true, details) || testCase.dryRun, "requirement details should have a status: true")
			assert.True(t, testCase.exists || requested || testCase.dryRun, "requirement should have requested to create item")
		})
	}
}
