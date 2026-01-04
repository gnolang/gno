package requirements

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Draft Condition.
type draft struct{}

var _ Requirement = &draft{}

func (*draft) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(pr.GetDraft(), "This pull request is a draft", details)
}

func Draft() Requirement {
	return &draft{}
}
