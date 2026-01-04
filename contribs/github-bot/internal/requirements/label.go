package requirements

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// LabelAction controls what to do with the given label.
type LabelAction byte

const (
	// LabelApply will place the label on the PR if it doesn't exist.
	LabelApply = iota
	// LabelRemove will remove the label from the PR if it exists.
	LabelRemove
	// LabelIgnore always leaves the label on the PR as-is, without modifying it.
	LabelIgnore
)

// Label Requirement.
type label struct {
	gh     *client.GitHub
	name   string
	action LabelAction
}

var _ Requirement = &label{}

func (l *label) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("This label is applied to pull request: %s", l.name)

	found := false
	// Check if label was already applied to PR.
	for _, label := range pr.Labels {
		if l.name == label.GetName() {
			found = true
			break
		}
	}

	// If in a dry run, or no action expected, skip applying the label.
	if l.gh.DryRun ||
		l.action == LabelIgnore ||
		(l.action == LabelApply && found) ||
		(l.action == LabelRemove && !found) {
		return utils.AddStatusNode(found, detail, details)
	}

	switch l.action {
	case LabelApply:
		// If label not already applied, apply it.
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
	case LabelRemove:
		// If label not already applied, apply it.
		if _, err := l.gh.Client.Issues.RemoveLabelForIssue(
			l.gh.Ctx,
			l.gh.Owner,
			l.gh.Repo,
			pr.GetNumber(),
			l.name,
		); err != nil {
			l.gh.Logger.Errorf("Unable to remove label %s from PR %d: %v", l.name, pr.GetNumber(), err)
			return utils.AddStatusNode(true, detail, details)
		}
		return utils.AddStatusNode(false, detail, details)
	default:
		panic(fmt.Sprintf("invalid LabelAction value: %d", l.action))
	}
}

// Label asserts that the label with the given name is not applied to the PR.
//
// If it's not a dry run, the label will be applied to the PR.
func Label(gh *client.GitHub, name string, action LabelAction) Requirement {
	return &label{gh, name, action}
}
