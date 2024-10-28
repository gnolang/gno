package requirement

import (
	"fmt"

	"github.com/google/go-github/v66/github"
)

// And Requirement
type and struct {
	requirements []Requirement
}

var _ Requirement = &and{}

// Validate implements Requirement
func (a *and) Validate(pr *github.PullRequest) bool {
	for _, requirement := range a.requirements {
		if !requirement.Validate(pr) {
			return false
		}
	}

	return true
}

// GetText implements Requirement
func (a *and) GetText() string {
	text := fmt.Sprintf("(%s", a.requirements[0].GetText())
	for _, requirement := range a.requirements[1:] {
		text = fmt.Sprintf("%s AND %s", text, requirement.GetText())
	}

	return text + ")"
}

func And(requirements ...Requirement) Requirement {
	if len(requirements) < 2 {
		panic("You should pass at least 2 requirements to And()")
	}

	return &and{requirements}
}

// Or Requirement
type or struct {
	requirements []Requirement
}

var _ Requirement = &or{}

// Validate implements Requirement
func (o *or) Validate(pr *github.PullRequest) bool {
	for _, requirement := range o.requirements {
		if !requirement.Validate(pr) {
			return false
		}
	}

	return true
}

// GetText implements Requirement
func (o *or) GetText() string {
	text := fmt.Sprintf("(%s", o.requirements[0].GetText())
	for _, requirement := range o.requirements[1:] {
		text = fmt.Sprintf("%s OR %s", text, requirement.GetText())
	}

	return text + ")"
}

func Or(requirements ...Requirement) Requirement {
	if len(requirements) < 2 {
		panic("You should pass at least 2 requirements to Or()")
	}

	return &or{requirements}
}

// Not Requirement
type not struct {
	req Requirement
}

var _ Requirement = &not{}

// Validate implements Requirement
func (n *not) Validate(pr *github.PullRequest) bool {
	return !n.req.Validate(pr)
}

// GetText implements Requirement
func (n *not) GetText() string {
	return fmt.Sprintf("NOT %s", n.req.GetText())
}

func Not(req Requirement) Requirement {
	return &not{req}
}
