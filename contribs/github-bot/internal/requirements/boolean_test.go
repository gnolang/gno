package requirements

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

func TestAnd(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name         string
		requirements []Requirement
		isSatisfied  bool
	}{
		{"and is true", []Requirement{Always(), Always()}, true},
		{"and is false", []Requirement{Always(), Always(), Never()}, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := And(testCase.requirements...)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}

func TestAndPanic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { And(Always()) }, "and constructor should panic if less than 2 conditions are provided")
}

func TestOr(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name         string
		requirements []Requirement
		isSatisfied  bool
	}{
		{"or is true", []Requirement{Never(), Always()}, true},
		{"or is false", []Requirement{Never(), Never(), Never()}, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := Or(testCase.requirements...)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}

func TestOrPanic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { Or(Always()) }, "or constructor should panic if less than 2 conditions are provided")
}

func TestNot(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		requirement Requirement
		isSatisfied bool
	}{
		{"not is true", Never(), true},
		{"not is false", Always(), false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := Not(testCase.requirement)

			assert.Equal(t, requirement.IsSatisfied(pr, details), testCase.isSatisfied, fmt.Sprintf("requirement should have a satisfied status: %t", testCase.isSatisfied))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isSatisfied, details), fmt.Sprintf("requirement details should have a status: %t", testCase.isSatisfied))
		})
	}
}

func TestIfCond(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		req         Requirement
		isSatisfied bool
	}{
		{"if always", If(Always()), true},
		{"if never", If(Never()), true},
		{"if always then always", If(Always()).Then(Always()), true},
		{"if never else always", If(Never()).Else(Always()), true},
		{"if always then never", If(Always()).Then(Never()), false},
		{"if never else never", If(Never()).Else(Never()), false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()

			actual := testCase.req.IsSatisfied(pr, details)
			assert.Equal(t, testCase.isSatisfied, actual,
				"requirement should have a satisfied status: %t", testCase.isSatisfied)
			assert.True(t,
				utils.TestNodeStatus(t, testCase.isSatisfied, details.(*treeprint.Node).Nodes[0]),
				"requirement details should have a status: %t", testCase.isSatisfied)
		})
	}
}

type reqFunc func(*github.PullRequest, treeprint.Tree) bool

func (r reqFunc) IsSatisfied(gh *github.PullRequest, details treeprint.Tree) bool {
	return r(gh, details)
}

func TestIfCond_ConditionalExecution(t *testing.T) {
	t.Run("executeThen", func(t *testing.T) {
		thenExec, elseExec := 0, 0
		If(Always()).
			Then(reqFunc(func(*github.PullRequest, treeprint.Tree) bool {
				thenExec++
				return true
			})).
			Else(reqFunc(func(*github.PullRequest, treeprint.Tree) bool {
				elseExec++
				return true
			})).IsSatisfied(nil, treeprint.New())
		assert.Equal(t, 1, thenExec, "Then should be executed 1 time")
		assert.Equal(t, 0, elseExec, "Else should be executed 0 time")
	})
	t.Run("executeElse", func(t *testing.T) {
		thenExec, elseExec := 0, 0
		If(Never()).
			Then(reqFunc(func(*github.PullRequest, treeprint.Tree) bool {
				thenExec++
				return true
			})).
			Else(reqFunc(func(*github.PullRequest, treeprint.Tree) bool {
				elseExec++
				return true
			})).IsSatisfied(nil, treeprint.New())
		assert.Equal(t, 0, thenExec, "Then should be executed 0 time")
		assert.Equal(t, 1, elseExec, "Else should be executed 1 time")
	})
}

func TestIfCond_NoRepeats(t *testing.T) {
	assert.Panics(t, func() {
		If(Always()).Then(Always()).Then(Always())
	}, "two Then should panic")
	assert.Panics(t, func() {
		If(Always()).Else(Always()).Else(Always())
	}, "two Else should panic")
}
