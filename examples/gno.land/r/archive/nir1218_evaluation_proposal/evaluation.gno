package evaluation

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
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

func (e *Evaluation) AddContribution(pr *PullRequest, contributor address) (int, bool) {
	id := pr.Id()
	e.pullrequests.Set(ufmt.Sprintf("%d", id), pr)
	c := NewContribution(id, contributor)
	e.contributions.Set(ufmt.Sprintf("%d", id), c)
	return id, true
}

func (e *Evaluation) UpdateContribution(id int, status string) bool {
	c, exists := e.contributions.Get(ufmt.Sprintf("%d", id))
	if exists {
		contribtution := c.(*Contribution)
		return contribtution.UpdateStatus(status)
	}
	return false
}
