// This package is used to manage the validator set.
package validators

import (
	"std"

	"gno.land/p/gov/proposal"
)

type Change struct {
	Address std.Address
	Power   int
}

// NewExecutor wraps a changes closure proposal to emit a typed object (subscribed by tm2) only if it passes through a complete p/gov/proposal process.
func NewExecutor(changesFn []Change) proposal.Executor {
	if changesFn == nil {
		panic("changesFn should not be nil")
	}

	// certify that the changes are sent from the context of this realm.
	callback := func() {
		changes := changesFn()
		std.Emit(changes) // depends on std.Emit implementation
	}

	exec := proposal.NewExecutor(callback)
	return exec
}
