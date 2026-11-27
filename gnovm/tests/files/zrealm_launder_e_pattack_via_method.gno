// PKGPATH: gno.land/r/laundere
package laundere

// A /p/ wrapper calling a /p/ method on another realm's object has no power
// over it: library code writes only objects of the current realm (R1).

// So the write below is refused.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadPtr())
	p := laundervictim.GetPtr()
	launderattack.TamperViaMethod(p, "pwnd-via-pattack-method")
	println("after: ", laundervictim.ReadPtr())
}

// Output:
// before: original-ptr

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): o<VPBlock(1,0)>.Field
