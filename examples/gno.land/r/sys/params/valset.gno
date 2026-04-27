package params

import (
	"strconv"

	prms "sys/params"
)

// Param keys read by gno.land/pkg/gnoland (EndBlocker).
// Keep in sync with gno.land/pkg/gnoland/app.go.
const (
	vmModulePrefix = "vm"

	// valsetDirtyKey signals the chain that proposed valset state
	// changed since the last EndBlocker. Realm sets true, EndBlocker
	// clears.
	valsetDirtyKey = "valset_dirty"

	// Two parallel typed lists per slot (new = proposed, prev = applied).
	// Length must match; address is derived from pubkey, so it isn't
	// stored separately.
	valsetNewPubKeysKey = "valset_new_pubkeys"
	valsetNewPowersKey  = "valset_new_powers"
	valsetPrevPubKeysKey = "valset_prev_pubkeys"
	valsetPrevPowersKey  = "valset_prev_powers"
)

// SetValsetProposal publishes a proposed valset for the chain to apply on
// the next EndBlocker. pubKeys[i] is paired with powers[i]. Power == 0
// removes the validator. Address is derived from pubkey on the chain side.
//
// realmPath is the publishing realm's package path (used as submodule);
// the chain reads from the realm configured in vm:p:valset_realm_path.
func SetValsetProposal(realmPath string, pubKeys []string, powers []int64) {
	if len(pubKeys) != len(powers) {
		panic("valset proposal: pubKeys and powers must have the same length")
	}
	prms.SetSysParamStrings(vmModulePrefix, realmPath, valsetNewPubKeysKey, pubKeys)
	prms.SetSysParamStrings(vmModulePrefix, realmPath, valsetNewPowersKey, formatPowers(powers))
	prms.SetSysParamBool(vmModulePrefix, realmPath, valsetDirtyKey, true)
}

// InitValsetPrev seeds an empty applied valset. Called once at realm init
// so that EndBlocker reads a valid (empty) prev list.
func InitValsetPrev(realmPath string) {
	prms.SetSysParamStrings(vmModulePrefix, realmPath, valsetPrevPubKeysKey, []string{})
	prms.SetSysParamStrings(vmModulePrefix, realmPath, valsetPrevPowersKey, []string{})
}

func formatPowers(powers []int64) []string {
	out := make([]string, len(powers))
	for i, p := range powers {
		out[i] = strconv.FormatInt(p, 10)
	}
	return out
}
