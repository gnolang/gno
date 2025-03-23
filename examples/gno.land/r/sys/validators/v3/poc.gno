package validators

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao/bridge"
)

// Keep in sync with vm/params.go
const valsetUpdatesKeyPrefix = "valset_updates"

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
		// piece of logic). I've used the realm here.
		// I've also chosen to utilize the existing []string setter,
		// and not introduce the "any" struct type setter, to
		// avoid opening the pandora's box on arbitrary structures.
		//
		// Params are dropped in the following params path (realm-local):
		// vm:gno.land/r/sys/validators/v3:valset_updates_height
		// with "valset_updates:height" being the key.
		//
		// @moul, the height is part of the key, because I went with the premise that
		// we will only modify params from GovDAO proposals. An alternative is to always reuse
		// the same param (key -> valset_updates) and wipe it in Endlocker, but this breaks the principle
		// that params should _only_ be updated through the GovDAO (in Gno).
		std.SetParamStrings(
			ufmt.Sprintf("%s_%d", valsetUpdatesKeyPrefix, std.ChainHeight()),
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

// Render displays the valset changes saved in the params,
// depending on the specified height
func Render(path string) string {
	router := mux.NewRouter()

	router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		height := std.ChainHeight()

		res.Write(renderHeight(height))
	})

	router.HandleFunc("{height}", func(res *mux.ResponseWriter, req *mux.Request) {
		heightStr := req.GetVar("height")

		height, err := strconv.Atoi(heightStr)
		if err != nil {
			res.Write("Invalid block height")

			return
		}

		res.Write(renderHeight(height))
	})

	return router.Render(path)
}

// renderHeight returns the render changes for the specific height
func renderHeight(height int) string {
	var sb strings.Builder

	// Fetch the changes from the params
	changes, found := std.GetParamStrings(
		ufmt.Sprintf("%s_%d", valsetUpdatesKeyPrefix, height),
	)
	if !found {
		sb.WriteString("No changes for height " + height)

		return sb.String()
	}

	sb.WriteString("## Valset changes at " + height + "\n")
	for _, change := range changes {
		sb.WriteString(change)
		sb.WriteString("\n")
	}

	return sb.String()
}
