package validators

import (
	"errors"
	"time"

	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
)

var (
	lastValsetUpdate     time.Time                       // last time the valset was updated
	valsetUpdateCooldown = time.Duration(24 * time.Hour) // cooldown period between valset updates
)

const (
	errNoChangesProposed    = "no set changes proposed"
	errValsetUpdateCooldown = "valset update cooldown in effect"
	errTooManyChanges       = "too many changes proposed; must be less than 2/3 of current validator set"
)

// NewPropRequest creates a new proposal request that wraps a changes closure
// proposal. This wrapper is required to ensure the GovDAO Realm actually
// executed the callback.
func NewPropRequest(changesFn func() []validators.Validator, title, description string) dao.ProposalRequest {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func(cur realm) error {
		// Enforce cooldown: proposals can only update the validator set after the cooldown period.
		if time.Since(lastValsetUpdate) < valsetUpdateCooldown {
			return errors.New(errValsetUpdateCooldown)
		}

		// Get the current total voting power of the validator set.
		currentVotingPower := vp.TotalVotingPower()

		// Track total voting power removed and added in this proposal.
		var totalVotingPowerRemoved, totalVotingPowerAdded uint64

		// Iterate through each proposed validator change.
		for _, change := range changesFn() {
			if change.VotingPower == 0 {
				// If voting power is zero, this is a request to remove a validator.
				existingVal, err := vp.GetValidator(change.Address)
				if err != nil {
					// Cannot remove a validator that does not exist.
					return errors.New("cannot remove non-existing validator: " + change.Address.String())
				}
				// Accumulate the voting power being removed.
				totalVotingPowerRemoved += existingVal.VotingPower

				removeValidator(change.Address)
				continue
			}

			// Otherwise, this is an addition or update.
			// Accumulate the voting power being added.
			totalVotingPowerAdded += change.VotingPower
			addValidator(change)
		}

		// Enforce the maximum allowed change threshold (2/3 of current voting power).
		if currentVotingPower > 0 {
			maxChanges := (currentVotingPower * 2) / 3
			if totalVotingPowerAdded > maxChanges || totalVotingPowerRemoved > maxChanges {
				// Reject the proposal if too much voting power is changed.
				return errors.New(errTooManyChanges)
			}
		}

		// Update the last valset update time
		lastValsetUpdate = time.Now()
		return nil
	}

	e := dao.NewSimpleExecutor(callback, "")

	return dao.NewProposalRequest(title, description, e)
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
