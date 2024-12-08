package conditions

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// CreatedFromFork Condition.
type createdFromFork struct{}

var _ Condition = &createdFromFork{}

func (b *createdFromFork) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(
		pr.GetHead().GetRepo().GetFullName() != pr.GetBase().GetRepo().GetFullName(),
		fmt.Sprintf("The pull request was created from a fork (head branch repo: %s)", pr.GetHead().GetRepo().GetFullName()),
		details,
	)
}

func CreatedFromFork() Condition {
	return &createdFromFork{}
}
