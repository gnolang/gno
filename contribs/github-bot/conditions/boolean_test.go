package conditions

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"

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

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}

func TestAndPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("and constructor should panic if less than 2 conditions are provided")
		}
	}()

	And(Always()) // Only 1 condition provided
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

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}

func TestOrPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("and constructor should panic if less than 2 conditions are provided")
		}
	}()

	Or(Always()) // Only 1 condition provided
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

			if condition.IsMet(pr, details) != testCase.isMet {
				t.Errorf("condition should have a met status: %t", testCase.isMet)
			}
			if !utils.TestLastNodeStatus(t, testCase.isMet, details) {
				t.Errorf("condition details should have a status: %t", testCase.isMet)
			}
		})
	}
}
