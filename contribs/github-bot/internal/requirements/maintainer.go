package requirements

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// MaintainerCanModify Requirement.
type maintainerCanModify struct{}

var _ Requirement = &maintainerCanModify{}

func (a *maintainerCanModify) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(
		pr.GetMaintainerCanModify(),
		"Maintainer can modify this pull request",
		details,
	)
}

func MaintainerCanModify() Requirement {
	return &maintainerCanModify{}
}
