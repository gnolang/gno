package requirements

import (
	"context"
	"net/http"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/logger"
	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/migueleliasweb/go-github-mock/src/mock"

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
		{"empty label list with dry-run", "label", []*github.Label{}, true, false},
		{"label list contains exact match", "label", labels, false, true},
		{"label list contains prefix match", "^lab", labels, false, true},
		{"label list contains prefix doesn't match", "lab$", labels, false, false},
		{"label list contains prefix doesn't match with dry-run", "lab$", labels, true, false},
		{"label list contains suffix match", "bel$", labels, false, true},
		{"label list contains suffix match with dry-run", "bel$", labels, true, true},
		{"label list contains suffix doesn't match", "^bel", labels, false, false},
		{"label list contains suffix doesn't match with dry-run", "^bel", labels, true, false},
		{"label list doesn't contains match", "baleb", labels, false, false},
		{"label list doesn't contains match with dry-run", "baleb", labels, true, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			requested := false
			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/issues/0/labels",
						Method:  "GET", // It looks like this mock package doesn't support mocking POST requests
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

			if !requirement.IsSatisfied(pr, details) && !testCase.dryRun {
				t.Errorf("requirement should have a satisfied status: %t", true)
			}
			if !utils.TestLastNodeStatus(t, true, details) && !testCase.dryRun {
				t.Errorf("requirement details should have a status: %t", true)
			}
			if !testCase.exists && !requested && !testCase.dryRun {
				t.Errorf("requirement should have requested to create item")
			}
		})
	}
}
