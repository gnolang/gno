package condition

import (
	"github.com/google/go-github/v66/github"
)

type Condition interface {
	// Check if the Condition is met by this PR
	Validate(pr *github.PullRequest) bool

	// Get a text representation of this Condition
	GetText() string
}
