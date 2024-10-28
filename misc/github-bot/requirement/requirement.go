package requirement

import (
	"github.com/google/go-github/v66/github"
)

type Requirement interface {
	// Check if the Requirement is met by this PR
	Validate(pr *github.PullRequest) bool

	// Get a text representation of this Requirement
	GetText() string
}
