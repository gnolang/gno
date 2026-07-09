package evaluation

import "gno.land/p/nt/avl/v0"

type TallyResult struct {
	results avl.Tree
}

type Tally interface {
	Tally()
}
