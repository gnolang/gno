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

func TestFileChanged(t *testing.T) {
	t.Parallel()

	filenames := []*github.CommitFile{
		{Filename: github.String("foo")},
		{Filename: github.String("bar")},
		{Filename: github.String("baz")},
	}

	for _, testCase := range []struct {
		name    string
		pattern string
		files   []*github.CommitFile
		isMet   bool
	}{
		{"empty file list", "foo", []*github.CommitFile{}, false},
		{"file list contains exact match", "foo", filenames, true},
		{"file list contains prefix match", "^fo", filenames, true},
		{"file list contains prefix doesn't match", "^oo", filenames, false},
		{"file list contains suffix match", "oo$", filenames, true},
		{"file list contains suffix doesn't match", "fo$", filenames, false},
		{"file list doesn't contains match", "foobar", filenames, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockedHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.EndpointPattern{
						Pattern: "/repos/pulls/0/files",
						Method:  "GET",
					},
					testCase.files,
				),
			)

			gh := &client.GitHub{
				Client: github.NewClient(mockedHTTPClient),
				Ctx:    context.Background(),
				Logger: logger.NewNoopLogger(),
			}

			pr := &github.PullRequest{}
			details := treeprint.New()
			condition := FileChanged(gh, testCase.pattern)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
		})
	}
}
