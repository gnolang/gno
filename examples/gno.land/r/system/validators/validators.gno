// This package is used to manage the validator set.
package validators

import govdao "gno.land/r/gov/dao"

type Change struct {
	Address std.Address
	Power   int
}

func ProposeChanges(changesFn func() []Change) int {
	// XXX: require payment?
	// XXX: sanitize proposal

	callback := func() {
		changes := changesFn()
		std.Emit(changes)
	}

	idx := govdao.ProposeChanges(govdao.Proposal{
		Callback: callback,
		Comment:  comment,
	})
	return idx
}
