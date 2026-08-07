package grc20

// IDGenerator issues the short ids forming the trailing segment of Token.ID().
// Obtain one from the realm that owns the counter, typically a registry.
//
// The counter stays in that realm: a generator carrying it in a field could be
// dereferenced into a copy from any package, and the copy would replay the
// sequence — the identity collision this design exists to prevent. Holding the
// realm's function instead means every copy advances the same counter, so no
// two calls return the same id.
type IDGenerator struct {
	packagePath string        // owning realm, captured at construction
	nextID      func() string // the owning realm's counter, advanced per call
}

// NewIDGenerator returns a Generator bound to the caller's realm and backed by nextID.
// rlm must be the caller's own captured cur (asserted via IsCurrent), so PackagePath()
// is unforgeable and a registry can accept only the ids its own counter produced;
// the leading unused int keeps rlm out of first position so this stays a non-crossing helper.
//
// nextID must advance state the caller owns. A degenerate one harms only its own
// realm: the resulting tokens carry that realm as IdentifierPath(), so a registry
// accepting only its own ids rejects them.
func NewIDGenerator(_ int, rlm realm, nextID func() string) *IDGenerator {
	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}

	pkgPath := rlm.PkgPath()
	if pkgPath == "" {
		panic(ErrNotRealm)
	}

	if nextID == nil {
		panic(ErrNilIDFunc)
	}

	return &IDGenerator{packagePath: pkgPath, nextID: nextID}
}

// PackagePath returns the realm the generator is bound to.
func (g *IDGenerator) PackagePath() string { return g.packagePath }

// NextID advances the owning realm's counter and returns the new id.
func (g *IDGenerator) NextID() string {
	id := g.nextID()
	if !validID(id) {
		panic(ErrInvalidID)
	}

	return id
}

// validID reports whether id is a valid trailing segment. It shares validSymbol's
// charset because it is concatenated into Token.ID(): banning `.` `/` and
// whitespace keeps ids unambiguous for downstream parsers.
func validID(id string) bool {
	return validSlugText(id, MaxIDLen)
}
