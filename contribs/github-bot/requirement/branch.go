package requirement

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

// Pass this to UpToDateWith constructor to check the PR head branch
// against its base branch
const PR_BASE = "PR_BASE"

// UpToDateWith Requirement
type upToDateWith struct {
	gh   *client.GitHub
	base string
}

var _ Requirement = &author{}

func (u *upToDateWith) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	base := u.base
	if u.base == PR_BASE {
		base = pr.GetBase().GetRef()
	}
	head := pr.GetHead().GetRef()

	cmp, _, err := u.gh.Client.Repositories.CompareCommits(u.gh.Ctx, u.gh.Owner, u.gh.Repo, base, head, nil)
	if err != nil {
		u.gh.Logger.Errorf("Unable to compare head branch (%s) and base (%s): %v", head, base, err)
		return false
	}

	return utils.AddStatusNode(
		cmp.GetBehindBy() == 0,
		fmt.Sprintf(
			"Head branch (%s) is up to date with base (%s): behind by %d / ahead by %d",
			head,
			base,
			cmp.GetBehindBy(),
			cmp.GetAheadBy(),
		),
		details,
	)
}

func UpToDateWith(gh *client.GitHub, base string) Requirement {
	return &upToDateWith{gh, base}
}
