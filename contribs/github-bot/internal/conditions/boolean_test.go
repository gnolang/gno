package conditions

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
		name       string
		conditions []Condition
		isMet      bool
	}{
		{"and is true", []Condition{Always(), Always()}, true},
		{"and is false", []Condition{Always(), Always(), Never()}, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			condition := And(testCase.conditions...)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
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
		name       string
		conditions []Condition
		isMet      bool
	}{
		{"or is true", []Condition{Never(), Always()}, true},
		{"or is false", []Condition{Never(), Never(), Never()}, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			condition := Or(testCase.conditions...)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
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
		name      string
		condition Condition
		isMet     bool
	}{
		{"not is true", Never(), true},
		{"not is false", Always(), false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			condition := Not(testCase.condition)

			assert.Equal(t, condition.IsMet(pr, details), testCase.isMet, fmt.Sprintf("condition should have a met status: %t", testCase.isMet))
			assert.True(t, utils.TestLastNodeStatus(t, testCase.isMet, details), fmt.Sprintf("condition details should have a status: %t", testCase.isMet))
		})
	}
}
