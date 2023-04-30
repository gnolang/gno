package evaluation

import "gno.land/p/demo/avl"

type Evaluation struct {
	contibutions avl.Tree
	pullrequests avl.Tree
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

func (e Evaluation) AddContribution(pr PullRequest, address std.Address, criteria map[string]int32) (int, bool) {
	id := pr.Id()
	pullrequests.Set(id, pr)
	c := NewContribution(id, address, criteria)
	contibutions.Set(id, c)
	return id, true
}

func (e Evaluation) UpdateContribution(id string, status string) bool {
	c, found := contributions.Get(id)
	if found == true {
		c.Status(status)
		return true
	}
	return false
}

func (e Evaluation) EvaluateContribution(id string) (int64, bool) {
	c, found := contributions.Get(id)
	var total int64
	if found == true {
		total = c.Evaluate()
		return total, found
	}
	return int64(0), false
}
