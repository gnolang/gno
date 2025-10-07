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

		// Start with the current total voting power and allowed change threshold.
		currentVotingPower := vp.TotalVotingPower()

		var totalVotingPowerAdded uint64

		for _, change := range changesFn() {
			if change.VotingPower == 0 {
				// Remove validator: update voting power and threshold.
				existingVal, err := vp.GetValidator(change.Address)
				if err != nil {
					return errors.New("cannot remove non-existing validator: " + change.Address.String())
				}
				currentVotingPower -= existingVal.VotingPower

				removeValidator(change.Address)
				continue
			}

			// Add validator: accumulate added voting power.
			totalVotingPowerAdded += change.VotingPower
			addValidator(change)
		}

		maxVotingPowerChange := (currentVotingPower * 2) / 3
		// Prevent proposals that add too much voting power at once.
		if maxVotingPowerChange > 1 && totalVotingPowerAdded >= maxVotingPowerChange {
			return errors.New(errTooManyChanges)
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
