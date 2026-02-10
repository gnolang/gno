// PKGPATH: gno.land/r/crossrealm
package crossrealm

// This tests for realm finalization upon implicit (borrow) realm switches.
// This is like zrealm_crossrealm30[a,b,c].gno, but with cross().
// Unlike a,b,c, this file saves sb locally, but crosses into it(self).

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type StructA struct {
	A any
}

type StructB struct {
	B *StructA
}

func (sb *StructB) Set(cur realm, sa *StructA) struct{} {
	sb.B = sa
	return struct{}{}
}

var sb *StructB

func init() {
	sb = &StructB{}
}

func main() {
	sa := &StructA{}
	sb.Set(cross, sa) // attach sa now.

	// If sa was attached to sb (this realm),
	// it would not attach to crossrealm_b below,
	// so we wouldn't be able to mutate it directly.
	crossrealm_b.SetObject(cross, sb)
	sa.A = 1 // should succeed
	if sa.A != 1 {
		panic("expected sa.A to be 1")
	}
	println("Success: sa direct manipulation succeeded, meaning it was attached by cross(sb.Set)(sa).")
}

// Output:
// Success: sa direct manipulation succeeded, meaning it was attached by cross(sb.Set)(sa).
