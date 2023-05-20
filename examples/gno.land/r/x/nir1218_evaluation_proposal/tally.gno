package evaluation

import (
	"std"

	"gno.land/p/demo/avl"
)

type TallyResult struct {
	results avl.Tree
}

type Tally interface {
	Tally()
}
