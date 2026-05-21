// PKGPATH: gno.land/r/launderj
package launderj

// Attack J: NIL-VALUED RECEIVERS OF OTHER REFERENCE-UNDERLYING TYPES.
//
// Variants of Attack H/I where the receiver is a nil function-typed
// value, nil slice, or nil map of a defined type. Each has
// `GetFirstObject` returning nil — same Rule-2 anchor gap. Each
// successive call overwrites imm.Field; the test asserts the final
// value to confirm all three shapes write through.
//
// Same preconditions as Attack H: upstream documents the attack as
// succeeding (`pwnd-via-func`, `-slice`, `-map`). Our branch retains
// master's readonly-taint check, which catches the EvilFunc write on
// the first attempt — execution aborts before EvilSlice/EvilMap fire.
// We assert the first panic to lock the protection in (same rationale
// as zrealm_launder_h_pcallback.gno).

import (
	"gno.land/p/demo/tests/launderattack"
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
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,1)>.Field
