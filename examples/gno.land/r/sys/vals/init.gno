package vals

import (
	"gno.land/p/sys/vals/poc"
	"gno.land/p/sys/vals/types"
)

func init() {
	v = &vals{
		p:       poc.NewPoC(),
		changes: make([]types.Validator, 0),
	}
}
