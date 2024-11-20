package requirements

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"
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
