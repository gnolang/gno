package users

import (
	"strings"

	"gno.land/p/nt/bptree/v0"
)

// canonicalStore maps canonicalized full names to the original name
// that was registered. Keyed by the result of Canonicalize.
//
// Multiple controllers (namereg/v1, future registries, governance,
// genesis bootstrapping) all share this single store via RegisterUser
// and the bypass variant RegisterUserIgnoreCanonical. Cross-controller
// canonical-collision detection is uniform and atomic with the
// nameStore write.
var canonicalStore = bptree.NewBPTree32()

// Canonicalize returns the canonical form of a name. The substitutions
// collapse single-character visual confusables that arise across the
// allowed [a-z0-9] character set, and strip the three separators that
// can sneak between identical alphanumeric runs.
//
//   - {l, i, 1} → i
//   - {0, o}    → o
//   - {-, ., _} → stripped
//   - all other characters unchanged
//
// CONTRACT: stable. Future controllers that want to share this
// canonical namespace MUST use this exact function — do not roll your
// own. Adding new substitutions later is a breaking change because it
// would silently re-key existing entries in canonicalStore.
//
// Input contract: ASCII-only. r/sys/users.validateName already rejects
// non-ASCII at the registration boundary, so by the time a name reaches
// Canonicalize through the standard write path it is guaranteed to be
// ASCII. Direct callers from other realms must honor this contract;
// non-ASCII bytes are passed through as-is and will produce undefined
// collision behavior.
//
// Multi-char confusables (m↔rn, nn↔m, cl↔d) are NOT canonicalized.
// They require fixed-point substring substitution rounds, which is out
// of scope for the unified store.
//
// Pure: no state access. Safe to call from anywhere.
func Canonicalize(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch c {
		case 'l', '1':
			b.WriteByte('i')
		case '0':
			b.WriteByte('o')
		case '-', '.', '_':
			// strip
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
