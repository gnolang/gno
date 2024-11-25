package requirements

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Always Requirement.
type always struct{}

var _ Requirement = &always{}

func (*always) IsSatisfied(_ *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(true, "On every pull request", details)
}

func Always() Requirement {
	return &always{}
}

// Never Requirement.
type never struct{}

var _ Requirement = &never{}

func (*never) IsSatisfied(_ *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(false, "On no pull request", details)
}

func Never() Requirement {
	return &never{}
}
