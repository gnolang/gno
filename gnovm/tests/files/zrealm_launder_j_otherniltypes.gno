// PKGPATH: gno.land/r/launderj
package launderj

// Attack J: NIL-VALUED RECEIVERS OF OTHER REFERENCE-UNDERLYING TYPES.
//
// Variants of Attack H/I where the receiver is a nil function-typed
// value, nil slice, or nil map of a defined type. Each has
// `GetFirstObject` returning nil — same anchor gap.
//
// The Attack H/I fix's recvDeclaredTypePkgPath handles these
// uniformly: the defined type's PkgPath is the anchor target
// regardless of whether the underlying is primitive, pointer, func,
// slice, or map.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	// Each of these should fail at the readonly check.

	r1 := revive(func() {
		var nilFn launderattack.EvilFunc
		gp.UseMutator(imm, nilFn)
	})
	println("EvilFunc nil-recv blocked:", r1 != nil)

	r2 := revive(func() {
		var nilSlice launderattack.EvilSlice
		gp.UseMutator(imm, nilSlice)
	})
	println("EvilSlice nil-recv blocked:", r2 != nil)

	r3 := revive(func() {
		var nilMap launderattack.EvilMap
		gp.UseMutator(imm, nilMap)
	})
	println("EvilMap nil-recv blocked:", r3 != nil)

	println("imm.Field unchanged:", laundervictim.ReadImm())
}

// Output:
// EvilFunc nil-recv blocked: true
// EvilSlice nil-recv blocked: true
// EvilMap nil-recv blocked: true
// imm.Field unchanged: original-imm
