package zentasktic_core

type Workable interface {
	// restrict implementation of Workable to this realm
	assertWorkable()
}

type isWorkable struct {}

func (wt *WorkableTask) assertWorkable() {}

func (wp *WorkableProject) assertWorkable() {}

var _ Workable = &WorkableTask{}
var _ Workable = &WorkableProject{}
