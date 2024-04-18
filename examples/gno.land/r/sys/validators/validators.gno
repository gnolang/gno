// This package is used to manage the validator set.
package validators

import (
	"std"

	govdao "gno.land/r/gov/dao"
)

type Change struct {
	Address std.Address
	Power   int
}

func ProposeChanges(changesFn func() []Change) int {
	if changesFn == nil {
		panic("changesFn should not be nil")
	}

	callback := func() {
		changes := changesFn()
		std.Emit(changes) // depends on std.Emit implementation
	}

	idx := govdao.ProposeChanges(govdao.Proposal{
		Callback: callback,
		Comment:  comment,
	})
	return idx
}
