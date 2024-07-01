package vals

import (
	"gno.land/p/sys/vals/poa"
	"gno.land/p/sys/vals/types"
)

func init() {
	v = &vals{
		p:       poa.NewPoA(),
		changes: make([]types.Validator, 0),
	}
}
