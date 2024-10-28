package requirement

import (
	"bot/client"

	"github.com/google/go-github/v66/github"
)

// Checkbox Requirement
type checkbox struct {
	gh   *client.Github
	desc string
}

var _ Requirement = &checkbox{}

// GetText implements Requirement
func (c *checkbox) GetText() string {
	return ""
}

// Validate implements Requirement
func (c *checkbox) Validate(pr *github.PullRequest) bool {
	return false
}

func Checkbox(gh *client.Github, desc string) Requirement {
	return &checkbox{gh: gh, desc: desc}
}
