package conditions

import (
	"fmt"
	"regexp"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// BaseBranch Condition.
type baseBranch struct {
	pattern *regexp.Regexp
}

var _ Condition = &baseBranch{}

func (b *baseBranch) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(
		b.pattern.MatchString(pr.GetBase().GetRef()),
		fmt.Sprintf("The base branch matches this pattern: %s", b.pattern.String()),
		details,
	)
}

func BaseBranch(pattern string) Condition {
	return &baseBranch{pattern: regexp.MustCompile(pattern)}
}

// HeadBranch Condition.
type headBranch struct {
	pattern *regexp.Regexp
}

var _ Condition = &headBranch{}

func (h *headBranch) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(
		h.pattern.MatchString(pr.GetHead().GetRef()),
		fmt.Sprintf("The head branch matches this pattern: %s", h.pattern.String()),
		details,
	)
}

func HeadBranch(pattern string) Condition {
	return &headBranch{pattern: regexp.MustCompile(pattern)}
}
