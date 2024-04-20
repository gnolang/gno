// Package validators is used to manage the validator set.
package validators

import (
	"std"

	"gno.land/p/gov/proposal"
)

// Change represents a change in the validator set.
type Change struct {
	Address std.Address
	Power   int
}

// NewProposalExecutor creates a new executor that wraps a changes closure
// proposal. It emits a typed object (subscribed by tm2) only if it passes
// through a complete p/gov/proposal process.
func NewProposalExecutor(changesFn []Change) proposal.Executor {
	if changesFn == nil {
		panic("changesFn should not be nil")
	}

	// Certify that the changes are sent from the context of this realm.
	callback := func() {
		changes := changesFn()

		// XXX: Not final API. Depends on the std.Emit implementation.
		std.Emit(changes)
	}

	exec := proposal.NewExecutor(callback)
	return exec
}
