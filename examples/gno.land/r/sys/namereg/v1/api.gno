package namereg

import (
	"errors"
	"regexp"
	"strings"
)

// Tier-1 username format. Anchored.
//   - literal `nym-` prefix (4 chars)
//   - 5-13 lowercase ASCII letters (the alpha stem)
//   - exactly 3 trailing decimal digits
//
// Total length 12-20 chars. Distinct by length from `g1...` addresses
// which are always 40 chars.
const reNymFormat = `^nym-[a-z]{5,13}\d{3}$`

var reNym = regexp.MustCompile(reNymFormat)

// Reserved alpha-stem prefixes. Names whose stem starts with one of
// these are rejected at format-validation time. Both `gi` and `gl`
// are listed even though `Canonicalize` maps `gl` to `gi` — this lets
// the rejection surface as ErrReservedPrefix at the format layer
// rather than as ErrCanonicalCollision after the fact, which is a
// clearer error for users.
var reservedPrefixes = []string{"gno", "gi", "gl"}

// Exported error sentinels returned by ValidateNymFormat. Use
// errors.Is or direct equality; do not string-match.
var (
	ErrInvalidFormat      = errors.New("namereg: name must match nym-[a-z]{5,13}\\d{3}")
	ErrReservedPrefix     = errors.New("namereg: stem starts with a reserved prefix (gno/gi/gl)")
	ErrBlacklisted        = errors.New("namereg: stem matches a reserved role name")
	ErrCanonicalCollision = errors.New("namereg: canonical form collides with an existing name")
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

// IsStemTaken reports whether the given alpha stem (or any l↔i
// canonical-form variant of it) is already registered through this
// controller. Pass the raw stem; canonicalization is applied
// internally — there is no way to misuse this by forgetting to
// canonicalize first.
//
// Returns the original (non-canonical) stem as the first result so
// callers can render the existing name in collision errors.
func IsStemTaken(stem string) (existing string, taken bool) {
	v, ok := canonicalStore.Get(Canonicalize(stem))
	if !ok {
		return "", false
	}
	return v.(string), true
}

// ValidateNymFormat checks the regex, prefix-exclusion, and reserved-
// name rules in that order. Returns one of the exported sentinel
// errors per failure mode, or nil on success.
//
// Does NOT run the canonical-collision check — that requires the
// caller to provide context (the colliding name needs to be surfaced
// to the user) and is not a pure format check. Run IsStemTaken
// separately for collision detection.
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
