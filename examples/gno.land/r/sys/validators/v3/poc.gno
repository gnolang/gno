package validators

import (
	"std"
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao/bridge"
)

// Keep in sync with gno.land/pkg/gnoland/app.go
const (
	// newUpdatesAvailableKey is a flag indicating the chain valset should be updated.
	// Set by the contract, but reset by the chain (EndBlocker)
	newUpdatesAvailableKey = "new_updates_available"

	// valsetNewKey is the param that holds the new proposed valset. Set by the contract,
	// and read (but never modified) by the chain
	valsetNewKey = "valset_new"

	// valsetPrevKey is the param that holds the latest valset. Initially set by the contract (init),
	// but later only written by the chain (EndBlocker).
	valsetPrevKey = "valset_prev"
)

const errNoChangesProposed = "no set changes proposed"

// NewValsetChangeExecutor creates a new GovDAO executor for proposing valset changes
func NewValsetChangeExecutor(changesFn func() []validators.Validator) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		changes := changesFn()
		if len(changes) == 0 {
			panic(errNoChangesProposed)
		}

		// Keep track of params changes
		for _, change := range changes {
			if change.VotingPower == 0 {
				// This change request is to remove the validator.
				// Update the on-chain set representation
				removeValidator(change.Address)

				continue
			}

			// This change request is to add the validator.
			// Update the on-chain set representation
			addValidator(change)
		}

		// Save the change in the VM params //

		// Set the new updates available flag, indicating
		// the valset should be read from the valsetNewKey param
		std.SetParamBool(newUpdatesAvailableKey, true)

		// Set the newest valset after all changes have been applied
		std.SetParamStrings(valsetNewKey, serializeValsetChanges(vp.GetValidators()))

		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(callback)
}

// serializeValsetChanges prepares the valset serialization,
// which will be saved in the VM params, and processed within EndBlocker logic
func serializeValsetChanges(set []validators.Validator) []string {
	serializedSet := make([]string, 0, len(set))

	for _, entry := range set {
		serializedEntry := ufmt.Sprintf(
			"%s:%s:%d",
			entry.Address.String(),
			entry.PubKey,
			entry.VotingPower,
		)

		serializedSet = append(serializedSet, serializedEntry)
	}

	return serializedSet
}

// IsValidator returns a flag indicating if the given bech32 address
// is part of the validator set
func IsValidator(addr std.Address) bool {
	return vp.IsValidator(addr)
}

// GetValidator returns the typed validator
func GetValidator(addr std.Address) validators.Validator {
	if validator, err := vp.GetValidator(addr); err == nil {
		return validator
	}

	panic("validator not found")
}

// GetValidators returns the typed validator set
func GetValidators() []validators.Validator {
	return vp.GetValidators()
}

// Render displays the current valset as well as proposed changes
// for the next height
func Render(string) string {
	height := std.ChainHeight()

	var sb strings.Builder

	writeSection(&sb, "Valset changes", valsetNewKey, height)
	writeSection(&sb, "Valset", valsetPrevKey, height)

	return sb.String()
}

// writeSection renders a specific valset render section
func writeSection(sb *strings.Builder, title, key string, height int64) {
	sb.WriteString(ufmt.Sprintf("## %s at %d\n\n", title, height))

	entries, found := std.GetParamStrings(key)
	if found && len(entries) > 0 {
		for _, e := range entries {
			sb.WriteString(e)
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString(ufmt.Sprintf("No %s available\n", title))
	}

	sb.WriteString("\n")
}
