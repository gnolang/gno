package validators

import (
	"chain/params"
	"chain/runtime"
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
)

// Keep in sync with gno.land/pkg/gnoland/app.go
const (
	// newUpdatesAvailableKey is a flag indicating the chain valset should be updated.
	// Set by the contract, but reset by the chain (EndBlocker).
	newUpdatesAvailableKey = "new_updates_available"

	// valsetNewKey is the param that holds the new proposed valset. Set by the contract,
	// and read (but never modified) by the chain.
	valsetNewKey = "valset_new"

	// valsetPrevKey is the param that holds the latest applied valset. Initially set by
	// the contract (init), but later only written by the chain (EndBlocker).
	valsetPrevKey = "valset_prev"
)

const errNoChangesProposed = "no set changes proposed"

// NewValsetChangeExecutor creates a new GovDAO executor for proposing valset changes.
// changesFn is evaluated eagerly; the resulting slice is captured by the returned
// Executor. Evaluating lazily would close over the caller's ephemeral realm and
// fail to persist via dao.MustCreateProposal ("cannot persist function from the
// private realm").
func NewValsetChangeExecutor(changesFn func() []validators.Validator) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	changes := changesFn()
	if len(changes) == 0 {
		panic(errNoChangesProposed)
	}

	callback := func(cur realm) error {
		// Apply each change to the on-chain valset.
		for _, change := range changes {
			if change.VotingPower == 0 {
				// Voting power == 0 means remove the validator.
				removeValidator(change.Address)
				continue
			}

			// Otherwise, add or update the validator.
			addValidator(change)
		}

		// Propagate the new valset through the VM params keeper.
		params.SetBool(newUpdatesAvailableKey, true)
		params.SetStrings(valsetNewKey, serializeValset(vp.GetValidators()))

		return nil
	}

	return dao.NewSimpleExecutor(callback, "")
}

// serializeValset serializes the validator set for storage in the params keeper.
// Format: "<address>:<pubkey>:<voting-power>"
func serializeValset(valset []validators.Validator) []string {
	serialized := make([]string, 0, len(valset))

	for _, v := range valset {
		serialized = append(serialized, ufmt.Sprintf(
			"%s:%s:%d",
			v.Address.String(),
			v.PubKey,
			v.VotingPower,
		))
	}

	return serialized
}

// IsValidator returns true if the given address is part of the validator set.
func IsValidator(addr address) bool {
	return vp.IsValidator(addr)
}

// GetValidator returns the validator with the given address.
// Panics if the validator is not found.
func GetValidator(addr address) validators.Validator {
	v, err := vp.GetValidator(addr)
	if err != nil {
		panic("validator not found")
	}

	return v
}

// GetValidators returns the current validator set.
func GetValidators() []validators.Validator {
	return vp.GetValidators()
}

// Render displays the current validator set.
func Render(string) string {
	var (
		sb  strings.Builder
		h   = runtime.ChainHeight()
		set = GetValidators()
	)

	sb.WriteString(ufmt.Sprintf("## Valset at height %d\n\n", h))

	if len(set) == 0 {
		sb.WriteString("Valset is empty.\n")
		return sb.String()
	}

	for i, v := range set {
		sb.WriteString(ufmt.Sprintf(
			"- #%d: %s (%d)\n",
			i,
			v.Address.String(),
			v.VotingPower,
		))
	}

	return sb.String()
}
