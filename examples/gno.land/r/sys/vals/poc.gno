package vals

import (
	"std"

	"gno.land/p/gov/proposal"
	"gno.land/r/sys/vals/poc"
	"gno.land/r/sys/vals/types"
)

const daoPkgPath = "gno.land/r/gov/dao"

const (
	errNoChangesProposed = "no set changes proposed"
	errNotGovDAO         = "caller not govdao executor"
)

// vals is the wrapper for the validator set protocol
type vals struct {
	p       types.Protocol    // p is the underlying validator set protocol
	changes []types.Validator // changes are the set changes that happened between scrapes
}

// v holds the active on-chain validator set state
var v = &vals{
	p:       poc.NewPoC(),
	changes: make([]types.Validator, 0),
}

// NewPropExecutor creates a new executor that wraps a changes closure
// proposal. This wrapper is required to ensure the GovDAO Realm actually
// executed the callback.
//
// Concept adapted from:
// https://github.com/gnolang/gno/pull/1945
func NewPropExecutor(changesFn func() []types.Validator) proposal.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		// Make sure the GovDAO executor runs the valset changes
		assertGovDAOCaller()

		for _, change := range changesFn() {
			if change.VotingPower == 0 {
				// This change request is to remove the validator
				removeValidator(change.Address)

				continue
			}

			// This change request is to add the validator
			addValidator(change.Address, change.PubKey)
		}

		return nil
	}

	return proposal.NewExecutor(callback)
}

// assertGovDAOCaller verifies the caller is the GovDAO executor
func assertGovDAOCaller() {
	if std.PrevRealm().PkgPath() != daoPkgPath {
		panic(errNotGovDAO)
	}
}

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(address std.Address, pubKey string) {
	if val := v.p.AddValidator(address, pubKey); val != nil {
		// Validator added, note the change
		v.changes = append(v.changes, types.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: val.VotingPower,
		})
	}
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func removeValidator(address std.Address) {
	if val := v.p.RemoveValidator(address); val != nil {
		// Validator removed, note the change
		v.changes = append(v.changes, types.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		})
	}
}

// IsValidator returns a flag indicating if the given bech32 address
// is part of the validator set
func IsValidator(address string) bool {
	return v.p.IsValidator(std.Address(address))
}

// GetValidators returns the typed validator set
func GetValidators() []*types.Validator {
	return v.p.GetValidators()
}

// getChanges returns the validator changes stored on the realm
func getChanges() []types.Validator {
	// Construct the changes
	changes := make([]types.Validator, len(v.changes))
	copy(changes, v.changes)

	// Reset the changes set
	v.changes = v.changes[:0]

	return changes
}
