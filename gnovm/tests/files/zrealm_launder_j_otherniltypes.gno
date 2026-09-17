// PKGPATH: gno.land/r/launderj
package launderj

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Attack J: NIL-VALUED RECEIVERS OF OTHER REFERENCE-UNDERLYING TYPES.
//
// Variants of Attack H/I where the receiver is a nil function-typed
// value, nil slice, or nil map of a defined type. Each has
// `GetFirstObject` returning nil — same borrow rule #2 anchor gap. Each
// successive call overwrites imm.Field; the test asserts the final
// value to confirm all three shapes write through.
//
// Same preconditions as Attack H: exploitable only when victim
// logic data is /p/-declared.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	var nilFn launderattack.EvilFunc
	gp.UseMutator(imm, nilFn)
	println("after EvilFunc:", laundervictim.ReadImm())

	var nilSlice launderattack.EvilSlice
	gp.UseMutator(imm, nilSlice)
	println("after EvilSlice:", laundervictim.ReadImm())

	var nilMap launderattack.EvilMap
	gp.UseMutator(imm, nilMap)
	println("after EvilMap:", laundervictim.ReadImm())
}

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,1)>.Field
