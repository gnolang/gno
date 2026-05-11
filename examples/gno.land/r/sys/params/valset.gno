package params

import (
	"chain"
	"chain/runtime"
	"errors"
	"strconv"
	"strings"

	prms "sys/params"

	"gno.land/p/sys/validators"
)

// Param keys read by gno.land/pkg/gnoland (EndBlocker).
// Keep in sync with gno.land/pkg/gnoland/node_params.go.
// nodeModulePrefix is declared in halt.gno (same package).
const (
	valsetSubmodule = "valset"

	// dirty signals the chain that the proposed valset differs from the
	// current applied one. Realm sets true; EndBlocker clears.
	valsetDirtyKey = "dirty"

	// One []string per slot; each entry has the form "<pubkey>:<power>"
	// (bech32 pubkey + decimal power). Address is derived from pubkey
	// on the chain side and not stored.
	//
	//   proposed = v3's full target valset
	//   current  = chain-managed: the set that becomes ACTIVE AT H+2
	//              once the most recent EndBlock's updates apply.
	//              NOT necessarily the set actively signing the current
	//              block — see ABCI H+2 sequencing.
	valsetProposedKey = "proposed"
	valsetCurrentKey  = "current"

	// Only this realm may write valset:proposed and valset:dirty.
	// valset:current is chain-managed (see ctx-sentinel in node_params.go).
	valsetAuthorizedRealm = "gno.land/r/sys/validators/v3"
)

// SetValsetProposal publishes the realm's desired valset. Each entry is
// "<bech32-pubkey>:<decimal-power>"; power=0 removes the validator.
// The chain reads this on the next EndBlocker, diffs it against
// valset:current, and propagates the changes to consensus.
func SetValsetProposal(cur realm, entries []string) {
	assertValsetCaller()
	prms.SetSysParamStrings(nodeModulePrefix, valsetSubmodule, valsetProposedKey, entries)
	prms.SetSysParamBool(nodeModulePrefix, valsetSubmodule, valsetDirtyKey, true)
}

// GetValsetEntries returns the chain's authoritative committed
// validator set (the contents of valset:current). This is the
// V_{H+2} view — the set that will be active at H+2 once the most
// recent EndBlock's updates apply, NOT the set signing the current
// block. Callers that want "what v3 reports as the current
// validator set" — including the in-flight proposed set during
// the dirty window — should call GetValsetEffective instead.
func GetValsetEntries() []validators.Validator {
	return parseValsetSlot(valsetCurrentKey)
}

// ValsetDirty reports whether valset:proposed is awaiting EndBlocker.
// Realm callers MUST treat this as transient: the dirty flag is set
// by SetValsetProposal and cleared by the chain's EndBlocker (every
// block where dirty=true on entry exits with dirty=false).
func ValsetDirty() bool {
	d, _ := prms.GetSysParamBool(nodeModulePrefix, valsetSubmodule, valsetDirtyKey)
	return d
}

// GetValsetEffective returns the set that WILL be active at H+2:
// valset:proposed if dirty, else valset:current. Used by v3 so that
// (a) reads after a same-block proposal callback see that proposal's
// effects, and (b) sequential same-block proposals accumulate
// correctly on top of each other.
//
// Misuse warning: this exists for r/sys/validators/v3's internal
// reads. Other realms making "is X a validator" decisions should
// call v3.IsValidator, not this directly, so future changes to v3's
// read semantics propagate uniformly.
func GetValsetEffective() []validators.Validator {
	key := valsetCurrentKey
	if ValsetDirty() {
		key = valsetProposedKey
	}
	return parseValsetSlot(key)
}

func parseValsetSlot(key string) []validators.Validator {
	raw, _ := prms.GetSysParamStrings(nodeModulePrefix, valsetSubmodule, key)
	out := make([]validators.Validator, 0, len(raw))
	for _, e := range raw {
		v, err := parseEntry(e)
		if err != nil {
			panic("valset:" + key + " corrupted: " + err.Error())
		}
		out = append(out, v)
	}
	return out
}

// parseEntry splits "<bech32-pubkey>:<decimal-power>" and derives the
// validator address via the chain.PubKeyAddress native helper.
func parseEntry(entry string) (validators.Validator, error) {
	pkStr, pStr, ok := strings.Cut(entry, ":")
	if !ok {
		return validators.Validator{}, errors.New("missing ':' separator in " + entry)
	}
	addr, err := chain.PubKeyAddress(pkStr)
	if err != nil {
		return validators.Validator{}, err
	}
	power, err := strconv.ParseUint(pStr, 10, 64)
	if err != nil {
		return validators.Validator{}, err
	}
	return validators.Validator{
		Address:     addr,
		PubKey:      pkStr,
		VotingPower: power,
	}, nil
}

func assertValsetCaller() {
	if runtime.PreviousRealm().PkgPath() != valsetAuthorizedRealm {
		panic("unauthorized: only " + valsetAuthorizedRealm + " may write valset params")
	}
}
