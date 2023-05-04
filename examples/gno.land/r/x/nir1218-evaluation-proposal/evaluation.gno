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

func (e *Evaluation) AddContribution(pr PullRequest, address std.Address, criteria map[string]int32) (int, bool) {
	id := pr.Id()
	e.pullrequests.Set(ufmt.Sprintf("%d", id), pr)
	c := NewContribution(id, address, criteria)
	e.contributions.Set(ufmt.Sprintf("%d", id), c)
	return id, true
}

func (e *Evaluation) UpdateContribution(id int, status string) bool {
	c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
	if found == true {
		contribtution := c.(*Contribution)
		contribtution.UpdateStatus(status)
		return true
	}
	return false
}

func (e Evaluation) EvaluateContribution(id int) (int64, bool) {
	c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
	var total int64
	if found == true {
		contribtution := c.(*Contribution)
		total = int64(contribtution.Evaluate())
		return total, found
	}
	return int64(0), false
}
