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

// GetValsetEntries returns the chain's authoritative validator set.
//
// IMPORTANT: this is the V_{H+2} view — the set that will be active
// at H+2 once the most recent EndBlock's updates apply, NOT the set
// actively signing the current block. Callers gating runtime decisions
// on "is X a validator right now" should be aware of the one-block
// transition window.
//
// Used by r/sys/validators/v3.init() to bootstrap its in-realm PoA
// from the genesis-seeded current set, so the first proposal diffs
// from truth (not from empty).
func GetValsetEntries() []validators.Validator {
	raw, _ := prms.GetSysParamStrings(nodeModulePrefix, valsetSubmodule, valsetCurrentKey)
	out := make([]validators.Validator, 0, len(raw))
	for _, e := range raw {
		v, err := parseEntry(e)
		if err != nil {
			panic("valset:current corrupted: " + err.Error())
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
