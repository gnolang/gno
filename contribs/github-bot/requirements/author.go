package requirements

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Author Requirement
type author struct {
	user string
}

var _ Requirement = &author{}

func (a *author) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	return utils.AddStatusNode(
		a.user == pr.GetUser().GetLogin(),
		fmt.Sprintf("Pull request author is user: %v", a.user),
		details,
	)
}

func Author(user string) Requirement {
	return &author{user: user}
}

// AuthorInTeam Requirement
type authorInTeam struct {
	gh   *client.GitHub
	team string
}

var _ Requirement = &authorInTeam{}

func (a *authorInTeam) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("Pull request author is a member of the team: %s", a.team)

	teamMembers, err := a.gh.ListTeamMembers(a.team)
	if err != nil {
		a.gh.Logger.Errorf("unable to check if author is in team %s: %v", a.team, err)
		return utils.AddStatusNode(false, detail, details)
	}

	for _, member := range teamMembers {
		if member.GetLogin() == pr.GetUser().GetLogin() {
			return utils.AddStatusNode(true, detail, details)
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

func AuthorInTeam(gh *client.GitHub, team string) Requirement {
	return &authorInTeam{gh: gh, team: team}
}
