package namereg

import (
	"errors"
	"regexp"
	"strings"
)

// Open Nym Tier username format. Anchored.
//   - literal `nym-` prefix (4 chars)
//   - 5-13 lowercase ASCII letters (the alpha stem)
//   - exactly 3 trailing decimal digits
//
// Total length 12-20 chars. Distinct by length from `g1...` addresses
// which are always 40 chars.
const reNymFormat = `^nym-[a-z]{5,13}\d{3}$`

var reNym = regexp.MustCompile(reNymFormat)

// Reserved alpha-stem prefixes. Names whose stem starts with one of
// these are rejected at format-validation time with ErrReservedPrefix
// (clearer than ErrCanonicalCollision after the fact).
//
// `gi` is intentionally NOT listed: in most rendering targets `i` is
// visually distinct enough from `1`/`l` that legitimate `gi*` names
// (giggles, gimbal, gift, etc.) should remain registerable. Phishing
// protection for the visual class is still enforced by canonical-
// collision detection in r/sys/users — once any `gi*` or `gl*` name
// is registered, all variants under the {l,i,1}→i canonicalization
// collide.
//
// `gl` and `g1` remain listed because they're more visually
// confusable with the bech32 address prefix `g1`. `g1` itself is
// unreachable through the alpha-only stem regex; defense-in-depth
// for any future regex relaxation.
var reservedPrefixes = []string{"gl", "g1", "gno", "atom", "atone", "photon", "cosmos"}

// Exported error sentinels returned by ValidateNymFormat. Use
// errors.Is or direct equality; do not string-match.
//
// ErrReservedPrefix's message is built from reservedPrefixes at package
// init time so the surfaced list never drifts from the actual policy.
//
// Canonical-collision detection moved to r/sys/users in Option B.
// Consumers that previously caught namereg.ErrCanonicalCollision
// should switch to susers.ErrCanonicalCollision.
var (
	ErrInvalidFormat  = errors.New("namereg: name must match nym-[a-z]{5,13}\\d{3}")
	ErrReservedPrefix = errors.New("namereg: stem starts with a reserved prefix (" + strings.Join(reservedPrefixes, "/") + ")")
	ErrBlacklisted    = errors.New("namereg: stem matches a reserved role name")
)

// IsReserved reports whether the given alpha stem matches a reserved
// role name (with implicit `s`-suffix expansion). The check is
// canonicalized — so `vital1k`-style l-substituted variants of a
// reserved name are also caught. O(1) backed by `reservedSet` built
// in init().
func IsReserved(stem string) bool {
	_, found := reservedSet[Canonicalize(stem)]
	return found
}

// ValidateNymFormat checks the regex, prefix-exclusion, and reserved-
// name rules in that order. Returns one of the exported sentinel
// errors per failure mode, or nil on success.
//
// Does NOT run the canonical-collision check — that lives in r/sys/users
// (susers.IsCanonicalTaken or, atomically with the write, inside
// susers.RegisterUser).
func ValidateNymFormat(username string) error {
	if !reNym.MatchString(username) {
		return ErrInvalidFormat
	}

	// Stem is everything between `nym-` (4 chars) and the trailing
	// 3 digits. Regex guarantees 5..13 alpha chars in this slice.
	stem := username[4 : len(username)-3]

	for _, p := range reservedPrefixes {
		if strings.HasPrefix(stem, p) {
			return ErrReservedPrefix
		}
	}

	if IsReserved(stem) {
		return ErrBlacklisted
	}

	return nil
}

// IsPaused exposes the realm's pause flag for cross-controller
// coordination.
func IsPaused() bool {
	return paused
}
