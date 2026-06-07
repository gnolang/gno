// PKGPATH: gno.land/r/launderrh_conv_iface_box
package launderrh_conv_iface_box

// CONVERSION VECTOR 2: BOXING — IMPLICIT assignment to interface.
// `var box any = foreignBarePtr` does NOT go through doOpConvert
// in the explicit-cast sense; it goes through ordinary assignment.
// Question: does N_Readonly survive into the boxed value and re-emerge
// after type assertion, blocking the subsequent write?
//
// Hypothesis: yes — the readonly bit travels with the value through
// the box, the assertion produces a readonly pointer again, and the
// assignment-time readonly check fires.

import (
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/launderrvictim"
)

func try(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			println(label, "blocked: true")
		} else {
			println(label, "blocked: false")
		}
	}()
	fn()
}

func main(cur realm) {
	// 1. Implicit box (assignment to any), type-assert, write.
	try("1 var any = ptr; .(*Bare).F=", func() {
		w := launderrvictim.GetWithBarePtr()
		var box any = w.Inner
		box.(*launderpkg.Bare).Field = "pwnd-implicit-iface"
	})

	// 2. Box THROUGH a parameter-typed-any, write
	try("2 fn(any), assert, write", func() {
		writeViaAny(launderrvictim.GetWithBarePtr().Inner)
	})

	// 3. Round-trip through fn-return any
	try("3 fn() any return, assert, write", func() {
		x := returnAsAny(launderrvictim.GetWithBarePtr().Inner)
		x.(*launderpkg.Bare).Field = "pwnd-return-any"
	})

	// 4. Box Bare VALUE (not pointer), then assert *Bare. Should fail
	//    type assertion (Bare != *Bare).
	try("4 val box, ptr-assert", func() {
		var box any = launderrvictim.GetBareSlice()[0]
		_ = box.(*launderpkg.Bare).Field
	})

	// 5. Box Bare VALUE, assert SAME value type (Bare), write the
	//    local copy. This should succeed (local) but NOT affect victim.
	try("5 val box, val-assert, local write", func() {
		var box any = launderrvictim.GetBareSlice()[0]
		b := box.(launderpkg.Bare)
		b.Field = "local-only"
		_ = b
	})

	println("--- final reads ---")
	println("ptr:", launderrvictim.ReadBarePtr())
	println("slice[0]:", launderrvictim.ReadBareSlice0())
}

func writeViaAny(x any) {
	t := x.(*launderpkg.Bare)
	t.Field = "pwnd-via-fn-any"
}

func returnAsAny(x any) any { return x }

// Output:
// 1 var any = ptr; .(*Bare).F= blocked: true
// 2 fn(any), assert, write blocked: true
// 3 fn() any return, assert, write blocked: true
// 4 val box, ptr-assert blocked: true
// 5 val box, val-assert, local write blocked: false
// --- final reads ---
// ptr: bare-ptr-orig
// slice[0]: slice0
