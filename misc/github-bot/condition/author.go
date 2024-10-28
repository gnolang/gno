package condition

import (
	"bot/client"
	"fmt"

	"github.com/google/go-github/v66/github"
)

// Author Condition
type author struct {
	user string
}

var _ Condition = &author{}

// GetText implements Condition
func (a *author) GetText() string {
	return fmt.Sprintf("Pull request author is user : %v", a.user)
}

// Validate implements Condition
func (a *author) Validate(pr *github.PullRequest) bool {
	return a.user == pr.GetUser().GetLogin()
}

func Author(user string) Condition {
	return &author{user: user}
}

// AuthorInTeam Condition
type authorInTeam struct {
	gh   *client.Github
	team string
}

var _ Condition = &authorInTeam{}

// GetText implements Condition
func (a *authorInTeam) GetText() string {
	return fmt.Sprintf("Pull request author is a member of the team : %s", a.team)
}

// Validate implements Condition
func (a *authorInTeam) Validate(pr *github.PullRequest) bool {
	for _, member := range a.gh.ListTeamMembers(a.team) {
		if member.GetLogin() == pr.GetUser().GetLogin() {
			return true
		}
	}

	return false
}

func AuthorInTeam(gh *client.Github, team string) Condition {
	return &authorInTeam{gh: gh, team: team}
}
