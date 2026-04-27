package params

import prms "sys/params"

// Param keys read by gno.land/pkg/gnoland (EndBlocker).
// Keep in sync with gno.land/pkg/gnoland/node_params.go.
// nodeModulePrefix is declared in halt.gno (same package).
const (
	valsetSubmodule = "valset"

	// dirty signals the chain that the proposed valset differs from the
	// previously applied one. Realm sets true; EndBlocker clears.
	valsetDirtyKey = "dirty"

	// One []string per slot; each entry has the form "<pubkey>:<power>"
	// (bech32 pubkey + decimal power; power=0 means removal). Address is
	// derived from pubkey on the chain side and not stored.
	//   new  = realm's desired valset
	//   prev = previously applied valset (what consensus is running)
	valsetNewKey  = "new"
	valsetPrevKey = "prev"
)

// SetValsetProposal publishes the realm's desired valset. Each entry is
// "<bech32-pubkey>:<decimal-power>"; power=0 removes the validator. The
// chain reads this on the next EndBlocker, diffs it against the prev
// set, and propagates the changes to consensus.
func SetValsetProposal(entries []string) {
	prms.SetSysParamStrings(nodeModulePrefix, valsetSubmodule, valsetNewKey, entries)
	prms.SetSysParamBool(nodeModulePrefix, valsetSubmodule, valsetDirtyKey, true)
}

// InitValsetPrev seeds an empty prev valset. Called once at realm init
// so that EndBlocker reads a valid (empty) list before any update.
func InitValsetPrev() {
	prms.SetSysParamStrings(nodeModulePrefix, valsetSubmodule, valsetPrevKey, []string{})
}
