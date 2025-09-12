// PKGPATH: gno.land/r/crossrealm
package crossrealm

// This tests for realm finalization upon implicit (borrow) realm switches.
// This is like zrealm_crossrealm30[b,c].gno, but with `return` for doOpReturnFromBlock.
// Assuming there are no defers:
// `return x` calls doOpReturn or doOpReturnAfterCopy if func results are named, while
// `return` calls doOpReturnFromBlock.

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type StructA struct {
	A any
}

type StructB struct {
	B *StructA
}

func (sb *StructB) Set(sa *StructA) {
	sb.B = sa
}

func init() {
	// save StructB{} in crossrealm_b.
	sb := &StructB{}
	crossrealm_b.SetObject(cross, sb)
}

func main() {
	sb := crossrealm_b.GetObject().(*StructB)
	sa := &StructA{}
	sb.Set(sa) // attach sa to crossrealm_b.

	sa.A = 1 // should fail
	panic("sa.A should not be mutable; it should be attached to crossrealm_b.")
}

// Error:
// invalid call: sb<VPBlock(1,0)>.Set(sa<VPBlock(1,1)>), method realm: RIDDC8307F7BBE97FDB07B2D0FA6907B8D67517C02A does not match receiver realm: RID46279D1E03ECB38B84822F9CAF584BD16913470D
