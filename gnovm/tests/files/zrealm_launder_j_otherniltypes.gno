// PKGPATH: gno.land/r/launderj
package launderj

// Explicit-entry rule: a /p/ method reached through an interface on a
// receiver that is not the realm's own runs view-only, so the write is
// refused. The analysis below was written before the explicit-entry rule.

// Attack J: NIL-VALUED RECEIVERS OF OTHER REFERENCE-UNDERLYING TYPES.
//
// Variants of Attack H/I where the receiver is a nil function-typed
// value, nil slice, or nil map of a defined type. Each has
// `GetFirstObject` returning nil — same borrow rule #2 anchor gap.
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

	// Each nil receiver reached through the Mutator interface runs view-only,
	// so the write aborts the transaction (one shape per run; EvilFunc here).
	var nilFn launderattack.EvilFunc
	gp.UseMutator(imm, nilFn)
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,1)>.Field
