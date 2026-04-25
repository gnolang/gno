package namereg

import (
	"strings"

	"gno.land/p/nt/bptree/v0"
)

// canonicalStore maps canonicalized alpha stems to the original stem
// that was registered. The original is preserved as the value so that
// when a future registration collides we can render the existing name
// to the user (better UX than just "this canonicalizes to something
// taken").
//
// Only the alpha stem is stored, not the full nym-XXX999 username.
// The prefix and digit suffix are constant across all Open Nym Tier
// registrations and contribute nothing to canonical-form distinction.
var canonicalStore = bptree.NewBPTree32()

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
var reservedSet map[string]struct{}

func init() {
	reservedSet = make(map[string]struct{}, len(reservedNames)*2)
	for _, n := range reservedNames {
		reservedSet[Canonicalize(n)] = struct{}{}
		reservedSet[Canonicalize(n+"s")] = struct{}{}
	}
}

// Canonicalize returns the canonical form of an alpha stem.
//
// CONTRACT: takes [a-z] input, returns [a-z] output. The function
// applies l→i substitution, which is the only single-char letter-
// letter confusable in [a-z] that survives the Open Nym Tier format's
// digit/letter zone separation (cross-script confusables like 0/o
// and 1/i,l can never occur in the alpha stem because digits and
// letters are in disjoint regex zones).
//
// STABLE API: any future controller that wants to share this
// canonical namespace MUST produce output in the same [a-z]
// codomain. If a future controller wants a different alphabet
// (uppercase, unicode), it must introduce its own canonicalization
// function under its own name — do not extend this one.
//
// Multi-char confusables (m↔rn, nn↔m, cl↔d) are NOT canonicalized.
// They require fixed-point substring substitution rounds, which is
// out of scope. The vv↔w confusable was considered and rejected as
// not worth the case for usernames.
//
// Pure: no state access. Safe to call from anywhere.
func Canonicalize(stem string) string {
	return strings.ReplaceAll(stem, "l", "i")
}
