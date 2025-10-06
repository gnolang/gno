// PKGPATH: gno.land/r/crossrealm
package crossrealm

// This tests for realm finalization upon implicit (borrow) realm switches.
// This is like zrealm_crossrealm30.gno, but with `return x` for doOpReturn.
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

func (sb *StructB) Set(sa *StructA) struct{} {
	sb.B = sa
	return struct{}{}
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
// cannot directly modify readonly tainted object (w/o method): sa<VPBlock(1,1)>.A
