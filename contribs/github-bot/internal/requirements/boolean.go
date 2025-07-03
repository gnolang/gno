package requirements

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// And Requirement.
type and struct {
	requirements []Requirement
}

var _ Requirement = &and{}

func (a *and) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	satisfied := utils.Success
	branch := details.AddBranch("")

	for _, requirement := range a.requirements {
		if !requirement.IsSatisfied(pr, branch) {
			satisfied = utils.Fail
			// We don't break here because we need to call IsSatisfied on all
			// requirements to populate the details tree.
		}
	}

	branch.SetValue(fmt.Sprintf("%s And", satisfied))

	return (satisfied == utils.Success)
}

func And(requirements ...Requirement) Requirement {
	if len(requirements) < 2 {
		panic("You should pass at least 2 requirements to And()")
	}

	return &and{requirements}
}

// Or Requirement.
type or struct {
	requirements []Requirement
}

var _ Requirement = &or{}

func (o *or) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	satisfied := utils.Fail
	branch := details.AddBranch("")

	for _, requirement := range o.requirements {
		if requirement.IsSatisfied(pr, branch) {
			satisfied = utils.Success
			// We don't break here because we need to call IsSatisfied on all
			// requirements to populate the details tree.
		}
	}

	branch.SetValue(fmt.Sprintf("%s Or", satisfied))

	return (satisfied == utils.Success)
}

func Or(requirements ...Requirement) Requirement {
	if len(requirements) < 2 {
		panic("You should pass at least 2 requirements to Or()")
	}

	return &or{requirements}
}

// Not Requirement.
type not struct {
	req Requirement
}

var _ Requirement = &not{}

func (n *not) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	satisfied := n.req.IsSatisfied(pr, details)
	node := details.FindLastNode()

	if satisfied {
		node.SetValue(fmt.Sprintf("%s Not (%s)", utils.Fail, node.(*treeprint.Node).Value.(string)))
	} else {
		node.SetValue(fmt.Sprintf("%s Not (%s)", utils.Success, node.(*treeprint.Node).Value.(string)))
	}

	return !satisfied
}

func Not(req Requirement) Requirement {
	return &not{req}
}

// IfCondition executes the condition, and based on the result then runs Then
// or Else.
type IfCondition struct {
	cond Requirement
	then Requirement
	els  Requirement
}

var _ Requirement = &IfCondition{}

func (i *IfCondition) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	if i.then == nil {
		i.then = Always()
	}
	ifBranch := details.AddBranch("")
	condBranch := ifBranch.AddBranch("")

	var (
		target     Requirement
		targetName string
	)

	if i.cond.IsSatisfied(pr, condBranch) {
		condBranch.SetValue(fmt.Sprintf("%s Condition", utils.Success))
		target, targetName = i.then, "Then"
	} else {
		condBranch.SetValue(fmt.Sprintf("%s Condition", utils.Fail))
		target, targetName = i.els, "Else"
	}

	targBranch := ifBranch.AddBranch("")
	if target == nil || target.IsSatisfied(pr, targBranch) {
		ifBranch.SetValue(fmt.Sprintf("%s If", utils.Success))
		targBranch.SetValue(fmt.Sprintf("%s %s", utils.Success, targetName))
		return true
	} else {
		ifBranch.SetValue(fmt.Sprintf("%s If", utils.Fail))
		targBranch.SetValue(fmt.Sprintf("%s %s", utils.Fail, targetName))
		return false
	}
}

// If returns a conditional requirement, which runs Then if cond evaluates
// successfully, or Else otherwise.
//
// Then / Else are optional, and always evaluate to true by default.
func If(cond Requirement) *IfCondition {
	return &IfCondition{cond: cond}
}

func (i *IfCondition) Then(then Requirement) *IfCondition {
	if i.then != nil {
		panic("'Then' is already set")
	}
	i.then = then
	return i
}

func (i *IfCondition) Else(els Requirement) *IfCondition {
	if i.els != nil {
		panic("'Else' is already set")
	}
	i.els = els
	return i
}
