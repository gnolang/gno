package identifier

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const selfPkgPath = "gno.land/p/onbloc/identifier"

// newTestGenerator constructs a Generator.
func newTestGenerator(_ int, rlm realm) (g *Generator) {
	func(cur realm) {
		g = NewGenerator(0, cur)
	}(cross(rlm))
	return
}

func TestNewGeneratorBindsPackagePath(cur realm, t *testing.T) {
	g := newTestGenerator(0, cur)
	urequire.False(t, g == nil, "generator should not be nil")
	uassert.Equal(t, selfPkgPath, g.PackagePath())
}

func TestNextIDShapeAndDeterminism(cur realm, t *testing.T) {
	g := newTestGenerator(0, cur)

	id1 := g.NextID()
	// 8 digest bytes encode to 13 Crockford base32 chars.
	uassert.Equal(t, 13, len(id1))

	// Same generator, next call: distinct code, counter advanced.
	id2 := g.NextID()
	uassert.NotEqual(t, id1, id2)

	// Determinism: a fresh generator bound to the same realm reproduces the
	// same sequence of codes (same package path, height, and counter inputs).
	h := newTestGenerator(0, cur)
	uassert.Equal(t, id1, h.NextID())
	uassert.Equal(t, id2, h.NextID())
}

func TestNextIDIsUnique(cur realm, t *testing.T) {
	g := newTestGenerator(0, cur)

	const n = 100
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		// Same block height throughout: uniqueness must come from the counter.
		id := g.NextID()
		_, dup := seen[id]
		urequire.False(t, dup, "NextID must never repeat: "+id)
		seen[id] = struct{}{}
	}
}

func TestNewGeneratorRejectsEOA(cur realm, t *testing.T) {
	// A user (EOA) realm has an empty package path.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("eoa")))
	uassert.PanicsWithMessage(t, cur, ErrNotRealm.Error(), func() {
		NewGenerator(0, cur)
	})
}
