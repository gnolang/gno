package vals

import (
	"gno.land/p/nt/poa"
	"gno.land/p/sys/vals"
)

func init() {
	v = &valset{
		p:       poa.NewPoA(),
		changes: make([]vals.Validator, 0),
	}
}
