package conditions

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Assignee Condition.
type assignee struct {
	user string
}

var _ Condition = &assignee{}

func (a *assignee) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("A pull request assignee is user: %s", a.user)

	for _, assignee := range pr.Assignees {
		if a.user == assignee.GetLogin() {
			return utils.AddStatusNode(true, detail, details)
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

func Assignee(user string) Condition {
	return &assignee{user: user}
}

// AssigneeInTeam Condition.
type assigneeInTeam struct {
	gh   *client.GitHub
	team string
}

var _ Condition = &assigneeInTeam{}

func (a *assigneeInTeam) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("A pull request assignee is a member of the team: %s", a.team)

	teamMembers, err := a.gh.ListTeamMembers(a.team)
	if err != nil {
		a.gh.Logger.Errorf("unable to check if assignee is in team %s: %v", a.team, err)
		return utils.AddStatusNode(false, detail, details)
	}

	for _, member := range teamMembers {
		for _, assignee := range pr.Assignees {
			if member.GetLogin() == assignee.GetLogin() {
				return utils.AddStatusNode(true, fmt.Sprintf("%s (member: %s)", detail, member.GetLogin()), details)
			}
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

func AssigneeInTeam(gh *client.GitHub, team string) Condition {
	return &assigneeInTeam{gh: gh, team: team}
}
