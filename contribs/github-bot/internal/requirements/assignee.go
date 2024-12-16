package requirements

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Assignee Requirement.
type assignee struct {
	gh   *client.GitHub
	user string
}

var _ Requirement = &assignee{}

func (a *assignee) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("This user is assigned to pull request: %s", a.user)

	// Check if user was already assigned to PR.
	for _, assignee := range pr.Assignees {
		if a.user == assignee.GetLogin() {
			return utils.AddStatusNode(true, detail, details)
		}
	}

	// If in a dry run, skip assigning the user.
	if a.gh.DryRun {
		return utils.AddStatusNode(false, detail, details)
	}

	// If user not already assigned, assign it.
	if _, _, err := a.gh.Client.Issues.AddAssignees(
		a.gh.Ctx,
		a.gh.Owner,
		a.gh.Repo,
		pr.GetNumber(),
		[]string{a.user},
	); err != nil {
		a.gh.Logger.Errorf("Unable to assign user %s to PR %d: %v", a.user, pr.GetNumber(), err)
		return utils.AddStatusNode(false, detail, details)
	}

	return utils.AddStatusNode(true, detail, details)
}

func Assignee(gh *client.GitHub, user string) Requirement {
	return &assignee{gh: gh, user: user}
}
