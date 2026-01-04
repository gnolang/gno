package conditions

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Author Condition.
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

// AuthorInTeam Condition.
type authorInTeam struct {
	gh   *client.GitHub
	team string
}

var _ Condition = &authorInTeam{}

func (a *authorInTeam) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
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

func AuthorInTeam(gh *client.GitHub, team string) Condition {
	return &authorInTeam{gh: gh, team: team}
}

type authorAssociationIs struct {
	assoc string
}

var _ Condition = &authorAssociationIs{}

func (a *authorAssociationIs) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("Pull request author has author_association: %q", a.assoc)

	return utils.AddStatusNode(pr.GetAuthorAssociation() == a.assoc, detail, details)
}

// AuthorAssociationIs asserts that the author of the PR has the given value for
// the GitHub "author association" field, on the PR.
//
// See https://docs.github.com/en/graphql/reference/enums#commentauthorassociation
// for a list of possible values and descriptions.
func AuthorAssociationIs(association string) Condition {
	return &authorAssociationIs{assoc: association}
}
