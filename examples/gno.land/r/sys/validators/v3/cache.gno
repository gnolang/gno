package validators

import (
	"chain"
	"chain/runtime"
	"sort"
	"strconv"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/sys/validators"
	sysparams "gno.land/r/sys/params"
)

// valopersRealmPath is the only realm allowed to refresh valoperCache
// or invoke RotateValoperSigningKey. Both auth checks below depend on
// being called via `cross` from a crossing function in valopers.
const valopersRealmPath = "gno.land/r/gnops/valopers"

// valoperCache mirrors the (operator -> current signing key) view from
// r/gnops/valopers. Written by valopers via NotifyValoperChanged. Read
// by future operator-keyed proposal flow (step 5).
//
// Pushing (valopers passes the values as args) rather than pulling
// (v3 imports valopers and reads them) — pulling would create an
// import cycle, since valopers already imports v3 for IsValidator.
var valoperCache = bptree.NewBPTree32()

type cacheEntry struct {
	SigningPubKey  string
	SigningAddress address
	KeepRunning    bool
}

// assertValopersCaller panics if the caller realm is not r/gnops/valopers.
// Per docs/resources/gno-interrealm.md, this check works only when (a)
// the host function is a crossing function (`cur realm`) and (b) it's
// invoked via `cross` from a crossing function in valopers. Then
// PreviousRealm() shifts exactly one frame to valopers. A user MsgCall
// would see PreviousRealm() == UserRealm (pkgpath ""); a third realm
// cross-call would see its own pkgpath. Either fails this check.
func assertValopersCaller() {
	caller := runtime.PreviousRealm().PkgPath()
	if caller != valopersRealmPath {
		panic("caller realm must be " + valopersRealmPath + ", got " + caller)
	}
}

// NotifyValoperChanged refreshes the cached entry for op. Auth: caller
// realm must be r/gnops/valopers.
//
// READ-ONLY against valopers: by design this function does not call
// back into valopers (no pull). Valopers pushes the current values in
// as args. This eliminates the confused-deputy class where v3 → valopers
// callbacks would make valopers see v3 as PreviousRealm.
func NotifyValoperChanged(cur realm, op address, signingPubKey string, signingAddress address, keepRunning bool) {
	assertValopersCaller()
	valoperCache.Set(op.String(), cacheEntry{
		SigningPubKey:  signingPubKey,
		SigningAddress: signingAddress,
		KeepRunning:    keepRunning,
	})
}

// RotateValoperSigningKey applies a signing-key rotation to the
// effective valset and publishes the new full set via sysparams.
// Auth: caller realm must be r/gnops/valopers.
//
// Body is read-modify-write against sysparams.GetValsetEffective so
// concurrent same-block writers (other rotations or GovDAO executors)
// accumulate instead of clobbering. Mirrors the executor pattern in
// validators.gno.
//
// Idempotent: if the rotating operator's old signing address is not
// currently in the effective valset (e.g., they were removed), the
// rotation is a no-op at the sysparams level — valopers' profile and
// signingRegistry already record the new key. Either replays cleanly.
//
// Emits ValoperRotated event with op + old/new addresses + height.
func RotateValoperSigningKey(cur realm, op address, oldPubKey, newPubKey string) {
	assertValopersCaller()

	oldAddr, err := chain.PubKeyAddress(oldPubKey)
	if err != nil {
		panic("invalid oldPubKey: " + err.Error())
	}
	newAddr, err := chain.PubKeyAddress(newPubKey)
	if err != nil {
		panic("invalid newPubKey: " + err.Error())
	}

	baseline := sysparams.GetValsetEffective()
	set := make(map[address]validators.Validator, len(baseline))
	for _, v := range baseline {
		set[v.Address] = v
	}

	// The rotating operator must currently be in the active set; if
	// not (operator removed before rotating), nothing to publish.
	// Valopers-side state has already been updated regardless.
	prev, ok := set[oldAddr]
	if !ok {
		chain.Emit(
			"ValoperRotated",
			"op", op.String(),
			"oldAddr", oldAddr.String(),
			"newAddr", newAddr.String(),
			"height", strconv.FormatInt(runtime.ChainHeight(), 10),
			"applied", "false",
		)
		return
	}

	delete(set, oldAddr)
	set[newAddr] = validators.Validator{
		Address:     newAddr,
		PubKey:      newPubKey,
		VotingPower: prev.VotingPower,
	}

	// Defense-in-depth: a delete+insert in this branch always leaves
	// at least one entry (the freshly inserted newAddr), so an empty
	// set is unreachable today. Panic explicitly anyway: a future
	// refactor of this body that ends up publishing an empty set
	// would otherwise be silently swallowed by the EndBlocker
	// (which logs and clears dirty for empty publishes), masking
	// the regression.
	if len(set) == 0 {
		panic("rotation would empty the validator set; refused to keep consensus liveness")
	}

	entries := make([]string, 0, len(set))
	for _, v := range set {
		entries = append(entries, v.PubKey+":"+strconv.FormatUint(v.VotingPower, 10))
	}
	sort.Strings(entries)
	sysparams.SetValsetProposal(cross, entries)

	chain.Emit(
		"ValoperRotated",
		"op", op.String(),
		"oldAddr", oldAddr.String(),
		"newAddr", newAddr.String(),
		"height", strconv.FormatInt(runtime.ChainHeight(), 10),
		"applied", "true",
	)
}

// AssertGenesisValopersConsistent panics if any entry in valset:current
// (the seeded genesis valset) lacks a corresponding valoperCache profile
// whose SigningAddress matches.
//
// **Genesis-mode only.** The function refuses to run unless
// runtime.ChainHeight() == 0. This is the documented intended usage
// (last migration .jsonl tx, before any block has been produced) and
// also closes a post-genesis MsgCall DoS surface — without the guard,
// an attacker could pay gas to repeatedly invoke an O(N) iteration
// over valoperCache + valset:current after the chain is live.
//
// gnoland's InitChainer auto-runs this assertion at end of
// genesis-mode replay when GnoGenesisState.PastChainIDs is non-empty;
// failure aborts the boot unconditionally. valoper-seed and
// hand-crafted migration .jsonls do NOT need to emit the call
// themselves.
//
// Crossing function: callable via MsgCall (only at genesis-mode).
// Doesn't mutate state — pure invariant check. Inverse direction
// (every valoperCache entry must have a corresponding valset:current
// entry) is intentionally NOT checked: extra valoper profiles
// registered without immediate valset inclusion are a normal
// post-genesis state.
func AssertGenesisValopersConsistent(cur realm) {
	if runtime.ChainHeight() != 0 {
		panic("AssertGenesisValopersConsistent is only callable during genesis-mode replay (ChainHeight()==0)")
	}

	// Collect the signing addresses present in valoperCache.
	seen := map[string]bool{}
	valoperCache.Iterate("", "", func(_ string, raw any) bool {
		entry := raw.(cacheEntry)
		seen[entry.SigningAddress.String()] = true
		return false
	})

	// Every entry in valset:current must appear in seen.
	for _, v := range sysparams.GetValsetEntries() {
		if !seen[v.Address.String()] {
			panic("genesis-validator " + v.Address.String() + " has no corresponding valoper profile (signing address not in valoperCache)")
		}
	}
}
