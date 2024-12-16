package requirements

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
	"github.com/stretchr/testify/assert"
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

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}
