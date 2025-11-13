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
	trustLevelRatio      = trustRatio{numerator: 1, denominator: 3}
)

const (
	errNoChangesProposed    = "no set changes proposed"
	errValsetUpdateCooldown = "valset update cooldown in effect"
	errTooManyChanges       = "too many changes proposed; trust level violated"
	errInvalidTrustLevel    = "invalid trust level fraction"
)

type trustRatio struct {
	numerator   uint64
	denominator uint64
}

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

		prevTotalVotingPower := vp.TotalVotingPower()

		proposedValidators := vp.GetValidatorsMap()

		var addedVotingPower uint64

		// Gather all proposed validator changes.
		for _, change := range changesFn() {
			val, exists := proposedValidators[change.Address]
			if exists {
				val.VotingPower = change.VotingPower
				proposedValidators[change.Address] = val
			} else {
				addedVotingPower += change.VotingPower
			}

			if change.VotingPower == 0 {
				if !exists {
					return errors.New("cannot remove non-existing validator: " + change.Address.String())
				}

				removeValidator(change.Address)
				continue
			}

			addValidator(change)
		}

		var updatedVotingPower uint64
		for _, val := range proposedValidators {
			updatedVotingPower += val.VotingPower
		}

		// Guardrails:
		// 1. Cannot remove all legacy validators.
		// 2. Legacy validators must retain at least the configured trust level of the voting power in the new set.
		if prevTotalVotingPower > 0 && (updatedVotingPower == 0 || updatedVotingPower*trustLevelRatio.denominator < (addedVotingPower+updatedVotingPower)*trustLevelRatio.numerator) {
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

// GetTrustLevel returns the currently configured trust level ratio.
func GetTrustLevel() (uint64, uint64) {
	return trustLevelRatio.numerator, trustLevelRatio.denominator
}

// NewTrustLevelPropRequest creates a governance proposal to update the trust level ratio.
func NewTrustLevelPropRequest(numerator, denominator uint64, title, description string) dao.ProposalRequest {
	if err := validateTrustLevelInputs(numerator, denominator); err != nil {
		panic(err.Error())
	}

	callback := func(cur realm) error {
		return setTrustLevel(numerator, denominator)
	}

	e := dao.NewSimpleExecutor(callback, "")

	return dao.NewProposalRequest(title, description, e)
}

func setTrustLevel(numerator, denominator uint64) error {
	if err := validateTrustLevelInputs(numerator, denominator); err != nil {
		return err
	}

	trustLevelRatio = trustRatio{numerator: numerator, denominator: denominator}
	return nil
}

func validateTrustLevelInputs(numerator, denominator uint64) error {
	if denominator == 0 || numerator == 0 || numerator > denominator {
		return errors.New(errInvalidTrustLevel)
	}
	return nil
}
