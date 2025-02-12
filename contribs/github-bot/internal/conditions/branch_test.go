package conditions

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

func TestHeadBaseBranch(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name    string
		pattern string
		base    string
		isMet   bool
	}{
		{"perfectly match", "base", "base", true},
		{"prefix match", "^dev/", "dev/test-bot", true},
		{"prefix doesn't match", "^/test-bot", "dev/test-bot", false},
		{"suffix match", "/test-bot$", "dev/test-bot", true},
		{"suffix doesn't match", "dev/$", "dev/test-bot", false},
		{"doesn't match", "base", "notatall", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{
				Base: &github.PullRequestBranch{Ref: github.String(testCase.base)},
				Head: &github.PullRequestBranch{Ref: github.String(testCase.base)},
			}
			conditions := []Condition{
				BaseBranch(testCase.pattern),
				HeadBranch(testCase.pattern),
			}

			for _, condition := range conditions {
				details := treeprint.New()
				assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
				assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
			}
		})
	}
}
