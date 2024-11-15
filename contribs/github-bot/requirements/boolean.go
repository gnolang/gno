package requirements

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// And Requirement
type and struct {
	requirements []Requirement
}

var _ Requirement = &and{}

func (a *and) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	satisfied := true
	branch := details.AddBranch("")

	for _, requirement := range a.requirements {
		if !requirement.IsSatisfied(pr, branch) {
			satisfied = false
		}
	}

	if satisfied {
		branch.SetValue(fmt.Sprintf("%s And", utils.StatusSuccess))
	} else {
		branch.SetValue(fmt.Sprintf("%s And", utils.StatusFail))
	}

	return satisfied
}

func And(requirements ...Requirement) Requirement {
	if len(requirements) < 2 {
		panic("You should pass at least 2 requirements to And()")
	}

	return &and{requirements}
}

// Or Requirement
type or struct {
	requirements []Requirement
}

var _ Requirement = &or{}

func (o *or) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	satisfied := false
	branch := details.AddBranch("")

	for _, requirement := range o.requirements {
		if requirement.IsSatisfied(pr, branch) {
			satisfied = true
		}
	}

	if satisfied {
		branch.SetValue(fmt.Sprintf("%s Or", utils.StatusSuccess))
	} else {
		branch.SetValue(fmt.Sprintf("%s Or", utils.StatusFail))
	}

	return satisfied
}

func Or(requirements ...Requirement) Requirement {
	if len(requirements) < 2 {
		panic("You should pass at least 2 requirements to Or()")
	}

	return &or{requirements}
}

// Not Requirement
type not struct {
	req Requirement
}

var _ Requirement = &not{}

func (n *not) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	satisfied := n.req.IsSatisfied(pr, details)
	node := details.FindLastNode()

	if satisfied {
		node.SetValue(fmt.Sprintf("%s Not (%s)", utils.StatusFail, node.(*treeprint.Node).Value.(string)))
	} else {
		node.SetValue(fmt.Sprintf("%s Not (%s)", utils.StatusSuccess, node.(*treeprint.Node).Value.(string)))
	}

	return !satisfied
}

func Not(req Requirement) Requirement {
	return &not{req}
}
