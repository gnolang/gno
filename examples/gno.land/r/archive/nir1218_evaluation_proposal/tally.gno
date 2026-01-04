package evaluation

import "gno.land/p/nt/avl"

type TallyResult struct {
	results avl.Tree
}

type Tally interface {
	Tally()
}
