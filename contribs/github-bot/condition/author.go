package condition

import (
	"bot/client"
	"bot/utils"
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

// Author Condition
type author struct {
	user string
}

var _ Condition = &author{}

func (a *author) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(
		a.user == pr.GetUser().GetLogin(),
		fmt.Sprintf("Pull request author is user: %v", a.user),
		details,
	)
}

func Author(user string) Condition {
	return &author{user: user}
}

// AuthorInTeam Condition
type authorInTeam struct {
	gh   *client.GitHub
	team string
}

var _ Condition = &authorInTeam{}

func (a *authorInTeam) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("Pull request author is a member of the team: %s", a.team)

	for _, member := range a.gh.ListTeamMembers(a.team) {
		if member.GetLogin() == pr.GetUser().GetLogin() {
			return utils.AddStatusNode(true, detail, details)
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

func AuthorInTeam(gh *client.GitHub, team string) Condition {
	return &authorInTeam{gh: gh, team: team}
}
