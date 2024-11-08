package condition

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v66/github"
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
		isMet   bool
	}{
		{"empty label list", "label", []*github.Label{}, false},
		{"label list contains exact match", "label", labels, true},
		{"label list contains prefix match", "^lab", labels, true},
		{"label list contains prefix doesn't match", "lab$", labels, false},
		{"label list contains suffix match", "bel$", labels, true},
		{"label list contains suffix doesn't match", "^bel", labels, false},
		{"label list doesn't contains match", "baleb", labels, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{Labels: testCase.labels}
			details := treeprint.New()
			condition := Label(testCase.pattern)

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}
