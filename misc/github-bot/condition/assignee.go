package condition

import (
	"bot/client"
	"fmt"

	"github.com/google/go-github/v66/github"
)

// Assignee Condition
type assignee struct {
	user string
}

var _ Condition = &assignee{}

// GetText implements Condition
func (a *assignee) GetText() string {
	return fmt.Sprintf("A pull request assignee is user : %s", a.user)
}

// Validate implements Condition
func (a *assignee) Validate(pr *github.PullRequest) bool {
	for _, assignee := range pr.Assignees {
		if a.user == assignee.GetLogin() {
			return true
		}
	}
	return false
}

func Assignee(user string) Condition {
	return &assignee{user: user}
}

// AssigneeInTeam Condition
type assigneeInTeam struct {
	gh   *client.Github
	team string
}

var _ Condition = &assigneeInTeam{}

// GetText implements Condition
func (a *assigneeInTeam) GetText() string {
	return fmt.Sprintf("A pull request assignee is a member of the team : %s", a.team)
}

// Validate implements Condition
func (a *assigneeInTeam) Validate(pr *github.PullRequest) bool {
	for _, member := range a.gh.ListTeamMembers(a.team) {
		for _, assignee := range pr.Assignees {
			if member.GetLogin() == assignee.GetLogin() {
				return true
			}
		}
	}

	return false
}

func AssigneeInTeam(gh *client.Github, team string) Condition {
	return &assigneeInTeam{gh: gh, team: team}
}
