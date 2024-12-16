package conditions

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
	"github.com/stretchr/testify/assert"
	"github.com/xlab/treeprint"
)

func TestDraft(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		isMet bool
	}{
		{"draft is true", true},
		{"draft is false", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{Draft: &testCase.isMet}
			details := treeprint.New()
			condition := Draft()

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
		})
	}
}
