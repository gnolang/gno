package requirements

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

func TestUpToDateWith(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		behind      int
		ahead       int
		isSatisfied bool
	}{
		{"up-to-date without commit ahead", 0, 0, true},
		{"up-to-date with commits ahead", 0, 3, true},
		{"not up-to-date with commits behind", 3, 0, false},
		{"not up-to-date with commits behind and ahead", 3, 3, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: "/repos/compare/base...",
						Method:  "GET",
					},
					github.CommitsComparison{
						AheadBy:  &testCase.ahead,
						BehindBy: &testCase.behind,
					},
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := UpToDateWith(gh, "base")

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}
