package validators

import (
	"std"
	prms "sys/params"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao/bridge"
)

const (
	vmModulePrefix         = "vm"
	valsetUpdatesKeyPrefix = "valset_updates"
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
		var (
			height     = std.ChainHeight()
			setChanges = make([]string, 0, len(changes))
		)

		for _, change := range changes {
			// Save the change for the VM params
			setChanges = append(
				setChanges,
				serializeValsetChange(height, change),
			)

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

		// Save the change in the VM params.
		// TODO: @moul, I'm not sure if this is the correct
		// place to dump the valset updates, given our
		// consensus module does not actually have params
		// (nor do I think it should, since it's a deterministic
		// piece of logic). I've used the VM module here.
		// I've also chosen to utilize the existing []string setter,
		// and not introduce the "any" struct type setter, to
		// avoid opening the pandora's box on arbitrary structures.
		//
		// Params are dropped in the following params path:
		// vm:p:valset_updates:height
		// with "valset_updates:height" being the key.
		// @moul, the height is part of the key, because I went with the premise that
		// we will only modify params from GovDAO proposals. An alternative is to always reuse
		// the same param (key -> valset_updates) and wipe it in Endlocker, but this breaks the principle
		// that params are _only_ updated through the GovDAO (in Gno).
		prms.SetSysParamStrings(
			vmModulePrefix,
			"p",
			ufmt.Sprintf("%s:%d", valsetUpdatesKeyPrefix, std.ChainHeight()),
			changes,
		)

		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(callback)
}

// serializeValsetChange prepares the valset change serialization,
// which will be saved in the VM params, and processed within EndBlocker logic
func serializeValsetChange(validator validators.Validator) string {
	// <address>:<pub-key>:<voting-power>
	// voting power == 0 => validator removal
	// voting power != 0 => validator power update / validator addition
	return ufmt.Sprintf(
		"%s:%s:%d",
		validator.Address.String(),
		validator.PubKey,
		validator.VotingPower,
	)
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
