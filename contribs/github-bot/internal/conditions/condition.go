package conditions

import (
	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

type Condition interface {
	// Check if the Condition is met and add the details
	// to the tree passed as a parameter.
	IsMet(pr *github.PullRequest, details treeprint.Tree) bool
}
