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
		{"file list contains prefix doesn't match", "fo$", filenames, false},
		{"file list contains suffix match", "oo$", filenames, true},
		{"file list contains suffix doesn't match", "^oo", filenames, false},
		{"file list doesn't contains match", "foobar", filenames, false},
	} {
		testCase := testCase
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

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}