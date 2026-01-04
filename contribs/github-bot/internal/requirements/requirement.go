package requirements

import (
	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

type Requirement interface {
	// Check if the Requirement is satisfied and add the detail
	// to the tree passed as a parameter.
	IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool
}
