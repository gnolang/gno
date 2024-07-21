package validators

import (
	"std"

	"gno.land/p/gov/proposal"
	"gno.land/p/sys/validators"
)

const daoPkgPath = "gno.land/r/gov/dao"

const (
	errNoChangesProposed = "no set changes proposed"
	errNotGovDAO         = "caller not govdao executor"
)

// NewPropExecutor creates a new executor that wraps a changes closure
// proposal. This wrapper is required to ensure the GovDAO Realm actually
// executed the callback.
//
// Concept adapted from:
// https://github.com/gnolang/gno/pull/1945
func NewPropExecutor(changesFn func() []validators.Validator) proposal.Executor {
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
			addValidator(change)
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

// IsValidator returns a flag indicating if the given bech32 address
// is part of the validator set
func IsValidator(addr std.Address) bool {
	return vp.IsValidator(addr)
}

// GetValidators returns the typed validator set
func GetValidators() []validators.Validator {
	return vp.GetValidators()
}
