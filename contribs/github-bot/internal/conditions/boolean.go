package conditions

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// And Condition.
type and struct {
	conditions []Condition
}

var _ Condition = &and{}

func (a *and) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	met := utils.Success
	branch := details.AddBranch("")

	for _, condition := range a.conditions {
		if !condition.IsMet(pr, branch) {
			met = utils.Fail
			// We don't break here because we need to call IsMet on all conditions
			// to populate the details tree.
		}
	}

	branch.SetValue(fmt.Sprintf("%s And", met))

	return (met == utils.Success)
}

func And(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("You should pass at least 2 conditions to And()")
	}

	return &and{conditions}
}

// Or Condition.
type or struct {
	conditions []Condition
}

var _ Condition = &or{}

func (o *or) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	met := utils.Fail
	branch := details.AddBranch("")

	for _, condition := range o.conditions {
		if condition.IsMet(pr, branch) {
			met = utils.Success
			// We don't break here because we need to call IsMet on all conditions
			// to populate the details tree.
		}
	}

	branch.SetValue(fmt.Sprintf("%s Or", met))

	return (met == utils.Success)
}

func Or(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("You should pass at least 2 conditions to Or()")
	}

	return &or{conditions}
}

// Not Condition.
type not struct {
	cond Condition
}

var _ Condition = &not{}

func (n *not) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	met := n.cond.IsMet(pr, details)
	node := details.FindLastNode()

	if met {
		node.SetValue(fmt.Sprintf("%s Not (%s)", utils.Fail, node.(*treeprint.Node).Value.(string)))
	} else {
		node.SetValue(fmt.Sprintf("%s Not (%s)", utils.Success, node.(*treeprint.Node).Value.(string)))
	}

	return !met
}

func Not(cond Condition) Condition {
	return &not{cond}
}
