package requirements

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"

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
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := And(testCase.requirements...)

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
		})
	}
}

func TestAndPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("and constructor should panic if less than 2 requirements are provided")
		}
	}()

	And(Always()) // Only 1 requirement provided
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
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := Or(testCase.requirements...)

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
		})
	}
}

func TestOrPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("and constructor should panic if less than 2 requirements are provided")
		}
	}()

	Or(Always()) // Only 1 requirement provided
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
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pr := &github.PullRequest{}
			details := treeprint.New()
			requirement := Not(testCase.requirement)

			if requirement.IsSatisfied(pr, details) != testCase.isSatisfied {
				t.Errorf("requirement should have a satisfied status: %t", testCase.isSatisfied)
			}
			if !utils.TestLastNodeStatus(t, testCase.isSatisfied, details) {
				t.Errorf("requirement details should have a status: %t", testCase.isSatisfied)
			}
		})
	}
}
