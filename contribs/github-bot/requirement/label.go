package requirement

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/client"
	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

// Label Requirement
type label struct {
	gh   *client.GitHub
	name string
}

var _ Requirement = &label{}

func (l *label) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("This label is applied to pull request: %s", l.name)

	// Check if label was already applied to PR
	for _, label := range pr.Labels {
		if l.name == label.GetName() {
			return utils.AddStatusNode(true, detail, details)
		}
	}

	// If in a dry run, skip applying the label
	if l.gh.DryRun {
		return utils.AddStatusNode(false, detail, details)
	}

	// If label not already applied, apply it
	if _, _, err := l.gh.Client.Issues.AddLabelsToIssue(
		l.gh.Ctx,
		l.gh.Owner,
		l.gh.Repo,
		pr.GetNumber(),
		[]string{l.name},
	); err != nil {
		l.gh.Logger.Errorf("Unable to add label %s to PR %d: %v", l.name, pr.GetNumber(), err)
		return utils.AddStatusNode(false, detail, details)
	}

	return utils.AddStatusNode(true, detail, details)
}

func Label(gh *client.GitHub, name string) Requirement {
	return &label{gh, name}
}
