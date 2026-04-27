package params

import (
	"strconv"

	prms "sys/params"
)

// Param keys read by gno.land/pkg/gnoland (EndBlocker).
// Keep in sync with gno.land/pkg/gnoland/app.go.
const (
	vmModulePrefix = "vm"
	vmParamSlot    = "p"

	// valsetDirtyKey signals the chain that the proposed valset differs
	// from the previously applied one. Realm sets true; EndBlocker clears.
	valsetDirtyKey = "valset_dirty"

	// One []string per slot; each entry has the form "<pubkey>:<power>"
	// (bech32 pubkey + decimal power; power=0 means removal). Address is
	// derived from pubkey on the chain side and not stored.
	//   new  = realm's desired valset
	//   prev = previously applied valset (what consensus is running)
	valsetNewKey  = "valset_new"
	valsetPrevKey = "valset_prev"
)

// SetValsetProposal publishes the realm's desired valset. The chain reads
// this on the next EndBlocker, diffs it against the prev set, and
// propagates the changes to consensus. Power == 0 removes the validator.
func SetValsetProposal(pubKeys []string, powers []int64) {
	if len(pubKeys) != len(powers) {
		panic("valset proposal: pubKeys and powers must have the same length")
	}
	entries := make([]string, len(pubKeys))
	for i, pk := range pubKeys {
		entries[i] = pk + ":" + strconv.FormatInt(powers[i], 10)
	}
	prms.SetSysParamStrings(vmModulePrefix, vmParamSlot, valsetNewKey, entries)
	prms.SetSysParamBool(vmModulePrefix, vmParamSlot, valsetDirtyKey, true)
}

// InitValsetPrev seeds an empty prev valset. Called once at realm init
// so that EndBlocker reads a valid (empty) list before any update.
func InitValsetPrev() {
	prms.SetSysParamStrings(vmModulePrefix, vmParamSlot, valsetPrevKey, []string{})
}
