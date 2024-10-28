package requirement

import (
	"bot/client"

	"github.com/google/go-github/v66/github"
)

// Assignee Requirement
type assignee struct {
	gh   *client.Github
	user string
}

var _ Requirement = &assignee{}

// GetText implements Requirement
func (a *assignee) GetText() string {
	return "TODO"
}

// Validate implements Requirement
func (a *assignee) Validate(pr *github.PullRequest) bool {
	// Check if user was already assigned to PR
	for _, assignee := range pr.Assignees {
		if a.user == assignee.GetLogin() {
			return true
		}
	}

	// If not, assign it
	if _, _, err := a.gh.Client.Issues.AddAssignees(
		a.gh.Ctx,
		a.gh.Owner,
		a.gh.Repo,
		pr.GetNumber(),
		[]string{a.user},
	); err != nil {
		a.gh.Logger.Errorf("Unable to assign user %s to PR %d : %v", a.user, pr.GetNumber(), err)
		return false
	}
	return true
}

func Assignee(gh *client.Github, user string) Requirement {
	return &assignee{gh: gh, user: user}
}
