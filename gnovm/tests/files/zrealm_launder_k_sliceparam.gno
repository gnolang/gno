// PKGPATH: gno.land/r/launderk
package launderk

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Attack K: same shape as H but the interface method takes a
// `[]*orig.Immutable` parameter (slice of foreign-/p/ pointers)
// instead of a bare `*orig.Immutable`. The slice element write
// inside the impl body commits the same way the bare-pointer write
// did. Documents that composite-parameter shapes don't add or
// remove anything from the H precondition.
//
// Same preconditions as Attack H: exploitable only when victim
// logic data is /p/-declared.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadImm())

	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	gp.UseSliceMutator(imm, launderattack.EvilSliceMutator(0))

	println("after: ", laundervictim.ReadImm())
}

// Output:
// before: original-imm

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): s<VPBlock(1,1)>[(const (0 int))].Field
