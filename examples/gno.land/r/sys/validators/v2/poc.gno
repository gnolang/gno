package validators

import (
	"std"

	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao/proxy"
)

const errNoChangesProposed = "no set changes proposed"

// NewPropRequest creates a new proposal request that wraps a changes closure
// proposal. This wrapper is required to ensure the GovDAO Realm actually
// executed the callback.
func NewPropRequest(changesFn func() []validators.Validator, description string) proxy.ProposalRequest {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func(proxy.Metadata) error {
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

	return proxy.ProposalRequest{
		Title:       "Validator Change Request",
		Description: description,
		Callback:    callback,
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
