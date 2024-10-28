package requirement

import (
	"bot/client"

	"github.com/google/go-github/v66/github"
)

// Label Requirement
type label struct {
	gh   *client.Github
	name string
}

var _ Requirement = &label{}

// Validate implements Requirement
func (l *label) Validate(pr *github.PullRequest) bool {
	// Check if label was already added to PR
	for _, label := range pr.Labels {
		if l.name == label.GetName() {
			return true
		}
	}

	// If not, add it
	if _, _, err := l.gh.Client.Issues.AddLabelsToIssue(
		l.gh.Ctx,
		l.gh.Owner,
		l.gh.Repo,
		pr.GetNumber(),
		[]string{l.name},
	); err != nil {
		l.gh.Logger.Errorf("Unable to add label %s to PR %d : %v", l.name, pr.GetNumber(), err)
		return false
	}
	return true
}

// GetText implements Requirement
func (l *label) GetText() string {
	return "TODO"
}

func Label(gh *client.Github, name string) Requirement {
	return &label{gh, name}
}
