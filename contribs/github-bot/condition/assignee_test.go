package condition

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

func TestAssignee(t *testing.T) {
	t.Parallel()

	assignees := []*github.User{
		{Login: github.String("notTheRightOne")},
		{Login: github.String("user")},
		{Login: github.String("anotherOne")},
	}

	for _, testCase := range []struct {
		name      string
		user      string
		assignees []*github.User
		isMet     bool
	}{
		{"empty assignee list", "user", []*github.User{}, false},
		{"assignee list contains user", "user", assignees, true},
		{"assignee list doesn't contain user", "user2", assignees, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{Assignees: testCase.assignees}
			details := treeprint.New()
			condition := Assignee(testCase.user)

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}
