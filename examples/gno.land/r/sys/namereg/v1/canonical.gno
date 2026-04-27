package namereg

import (
	susers "gno.land/r/sys/users"
)

// reservedSet is the runtime O(1) lookup for the role-name blacklist.
// Built in init() from reservedNames in blacklist.gno: each source
// entry contributes BOTH `Canonicalize(n)` and `Canonicalize(n+"s")`
// as keys, implementing the "implicit `s` suffix" rule documented on
// reservedNames.
//
// Why canonicalize the blacklist itself: validation canonicalizes the
// candidate stem before checking, so the comparison set must also be
// in canonical form. Otherwise a candidate like "vital1k" would
// canonicalize to "vitaiik" but the blacklist would contain only
// "vitalik" — the comparison would miss the match. Keeping both sides
// in canonical form makes the lookup exact.
//
// The blacklist remains namereg/v1-local because it is policy specific
// to the Open Nym Tier. Other controllers may have entirely different
// reserved-name policies, or none at all.
var reservedSet map[string]struct{}

func init() {
	reservedSet = make(map[string]struct{}, len(reservedNames)*2)
	for _, n := range reservedNames {
		reservedSet[Canonicalize(n)] = struct{}{}
		reservedSet[Canonicalize(n+"s")] = struct{}{}
	}
}

// Canonicalize is a delegating shim to r/sys/users.Canonicalize.
//
// HISTORY: namereg/v1 used to host its own per-stem canonical store and
// its own Canonicalize (l→i only). Option B unified the canonical lookup
// into r/sys/users keyed by full canonical name with broader
// substitutions ({l,i,1}→i, {0,o}→o, {-,.,_} stripped). This shim
// preserves the call-site name for local consumers (blacklist init,
// IsReserved) and any external consumer that imported the function
// from namereg/v1 before Option B.
//
// New code should call susers.Canonicalize directly.
func Canonicalize(s string) string {
	return susers.Canonicalize(s)
}
