// Package validators is used to manage the validator set.
package validators

import (
	"std"
	"strconv"

	"gno.land/p/gov/proposal"
)

var unappliedChanges = []Change{}

// Change represents a change in the validator set.
type Change struct {
	Address std.Address
	Power   int
}

// NewPropExecutor creates a new executor that wraps a changes closure
// proposal. It emits a typed object (subscribed by tm2) only if it passes
// through a complete p/gov/proposal process.
func NewPropExecutor(changesFn func() []Change) proposal.Executor {
	if changesFn == nil {
		panic("changesFn should not be nil")
	}

	// Certify that the changes are sent from the context of this realm.
	callback := func() error {
		newChanges := changesFn()

		// emit for external clients
		std.Emit("newChanges") // XXX: pass parameters

		// append to slice for gno.land
		unappliedChanges = append(unappliedChanges, newChanges...)
		return nil
	}

	exec := proposal.NewExecutor(callback)
	return exec
}

// this function is unexported and intended to be called by the chain.
func getAndResetChanges() []Change {
	cpy := unappliedChanges[:]
	unappliedChanges = []Change{}
	return cpy
}

func Render(_ string) string {
	if len(unappliedChanges) == 0 {
		return "No valset changes to apply."
	}

	output := "Valset changes to apply:\n"
	for _, change := range unappliedChanges {
		output += "- " + string(change.Address) + " (" + strconv.Itoa(change.Power) + ")\n"
	}
	return output
}
