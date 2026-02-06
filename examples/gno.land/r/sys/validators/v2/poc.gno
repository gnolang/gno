package validators

import (
	"strings"

	"gno.land/p/nt/ufmt"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
)

// NewPropRequest creates a new proposal request that wraps a changes closure
// proposal. This wrapper is required to ensure the GovDAO Realm actually
// executed the callback.
func NewPropRequest(changesFn func() []validators.Validator, title, description string) dao.ProposalRequest {
	if changesFn == nil {
		panic("no set changes proposed")
	}

	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}

	// Get the list of validators now to make sure the list
	// doesn't change during the lifetime of the proposal
	changes := changesFn()

	// Limit the number of validators to keep the description within a limit
	// that makes sense because there is not pagination of validators
	if len(changes) > 40 {
		panic("max number of allowed validators per proposal is 40")
	} else if len(changes) == 0 {
		panic("proposal requires at least one validator")
	}

	// List the validator addresses and the action to be taken for each one
	var desc strings.Builder
	desc.WriteString(description)
	if len(description) > 0 {
		desc.WriteString("\n\n")
	}

	desc.WriteString("## Validator Updates\n")
	for _, change := range changes {
		if change.VotingPower == 0 {
			desc.WriteString(ufmt.Sprintf("- %s: remove", change.Address))
		} else {
			desc.WriteString(ufmt.Sprintf("- %s: add", change.Address))
		}
	}

	callback := func(cur realm) error {
		for _, change := range changes {
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

	e := dao.NewSimpleExecutor(callback, "")

	return dao.NewProposalRequest(title, desc.String(), e)
}

// IsValidator returns a flag indicating if the given bech32 address
// is part of the validator set
func IsValidator(addr address) bool {
	return vp.IsValidator(addr)
}

// GetValidator returns the typed validator
func GetValidator(addr address) validators.Validator {
	if validator, err := vp.GetValidator(addr); err == nil {
		return validator
	}

	panic("validator not found")
}

// GetValidators returns the typed validator set
func GetValidators() []validators.Validator {
	return vp.GetValidators()
}
