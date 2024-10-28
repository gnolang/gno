package requirement

import (
	"github.com/google/go-github/v66/github"
)

// MaintainerCanModify Requirement
type maintainerCanModify struct{}

var _ Requirement = &maintainerCanModify{}

// GetText implements Requirement
func (a *maintainerCanModify) GetText() string {
	return "TODO"
}

// Validate implements Requirement
func (a *maintainerCanModify) Validate(pr *github.PullRequest) bool {
	return pr.GetMaintainerCanModify()
}

func MaintainerCanModify() Requirement {
	return &maintainerCanModify{}
}
