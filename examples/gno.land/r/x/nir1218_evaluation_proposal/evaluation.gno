package evaluation

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Evaluation struct {
	contributions avl.Tree
	pullrequests  avl.Tree
}

type Evaluator interface {
	Evaluate() Points
}

func NewEvalutaion() *Evaluation {
	e := &Evaluation{
		contributions: avl.Tree{},
		pullrequests:  avl.Tree{},
	}
	return e
}

func (e *Evaluation) AddContribution(pr *PullRequest, contributor std.Address) (contributionId int, ok bool) {
	id := pr.Id()
	e.pullrequests.Set(ufmt.Sprintf("%d", id), pr)
	c := NewContribution(id, contributor)
	e.contributions.Set(ufmt.Sprintf("%d", id), c)
	return id, true
}

func (e *Evaluation) UpdateContribution(id int, status string) bool {
	c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
	if found == true {
		contribtution := c.(*Contribution)
		return contribtution.UpdateStatus(status)
	}
	return false
}

// TODO
