package condition

import (
	"fmt"

	"github.com/google/go-github/v66/github"
)

// And Condition
type and struct {
	conditions []Condition
}

var _ Condition = &and{}

// Validate implements Condition
func (a *and) Validate(pr *github.PullRequest) bool {
	for _, condition := range a.conditions {
		if !condition.Validate(pr) {
			return false
		}
	}

	return true
}

// GetText implements Condition
func (a *and) GetText() string {
	text := fmt.Sprintf("(%s", a.conditions[0].GetText())
	for _, condition := range a.conditions[1:] {
		text = fmt.Sprintf("%s AND %s", text, condition.GetText())
	}

	return text + ")"
}

func And(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("You should pass at least 2 conditions to And()")
	}

	return &and{conditions}
}

// Or Condition
type or struct {
	conditions []Condition
}

var _ Condition = &or{}

// Validate implements Condition
func (o *or) Validate(pr *github.PullRequest) bool {
	for _, condition := range o.conditions {
		if condition.Validate(pr) {
			return true
		}
	}

	return false
}

// GetText implements Condition
func (o *or) GetText() string {
	text := fmt.Sprintf("(%s", o.conditions[0].GetText())
	for _, condition := range o.conditions[1:] {
		text = fmt.Sprintf("%s OR %s", text, condition.GetText())
	}

	return text + ")"
}

func Or(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("You should pass at least 2 conditions to Or()")
	}

	return &or{conditions}
}

// Not Condition
type not struct {
	cond Condition
}

var _ Condition = &not{}

// Validate implements Condition
func (n *not) Validate(pr *github.PullRequest) bool {
	return !n.cond.Validate(pr)
}

// GetText implements Condition
func (n *not) GetText() string {
	return fmt.Sprintf("NOT %s", n.cond.GetText())
}

func Not(cond Condition) Condition {
	return &not{cond}
}
