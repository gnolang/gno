package requirement

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

func TestMaintenerCanModify(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		isSatisfied bool
	}{
		{"modify is true", true},
		{"modify is false", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{MaintainerCanModify: &testCase.isSatisfied}
			details := treeprint.New()
			requirement := MaintainerCanModify()

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
		})
	}
}
