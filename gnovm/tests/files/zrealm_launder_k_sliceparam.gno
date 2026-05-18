// PKGPATH: gno.land/r/launderk
package launderk

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
	"gno.land/p/demo/tests/launderattack"
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
// after:  pwnd-via-slice-arg
