package condition

import (
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/xlab/treeprint"
)

// And Condition
type and struct {
	conditions []Condition
}

var _ Condition = &and{}

func (a *and) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	met := true
	branch := details.AddBranch("")

	for _, condition := range a.conditions {
		if !condition.IsMet(pr, branch) {
			met = false
		}
	}

	if met {
		branch.SetValue("🟢 And")
	} else {
		branch.SetValue("🔴 And")
	}

	return met
}

func And(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("You should pass at least 2 conditions to And()")
	}

	return &and{conditions}
}

// Or Condition
type or struct {
	conditions []Condition
}

var _ Condition = &or{}

func (o *or) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	met := false
	branch := details.AddBranch("")

	for _, condition := range o.conditions {
		if condition.IsMet(pr, branch) {
			met = true
		}
	}

	if met {
		branch.SetValue("🟢 Or")
	} else {
		branch.SetValue("🔴 Or")
	}

	return met
}

func Or(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("You should pass at least 2 conditions to Or()")
	}

	return &or{conditions}
}

// Not Condition
type not struct {
	cond Condition
}

var _ Condition = &not{}

func (n *not) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	met := n.cond.IsMet(pr, details)
	node := details.FindLastNode()

	if met {
		node.SetValue(fmt.Sprintf("🔴 Not (%s)", node.(*treeprint.Node).Value.(string)))
	} else {
		node.SetValue(fmt.Sprintf("🟢 Not (%s)", node.(*treeprint.Node).Value.(string)))
	}

	return !met
}

func Not(cond Condition) Condition {
	return &not{cond}
}
