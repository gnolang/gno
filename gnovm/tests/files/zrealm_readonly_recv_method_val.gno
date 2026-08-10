// PKGPATH: gno.land/r/test/readonlyrecvmethodval
package readonlyrecvmethodval

// Test C: value-receiver variant of zrealm_readonly_recv_method.gno.
// `v` is a TestRealmObject2 (not a pointer) obtained from an external
// /p/ package, so it carries the sticky N_Readonly bit. Then call a
// value-receiver method that writes `o2.Field = ...` on its local copy.
//
// Two possible outcomes:
//   (a) the sticky bit propagates onto the receiver-as-args[0] in the
//       new frame, and the in-method PopAsPointer panics, OR
//   (b) the bit is stripped during popCopyArgs (or the local-block TV
//       binding) and the method runs to completion; the local-copy
//       mutation is harmlessly discarded by value semantics, and the
//       caller's `v.Field` is unchanged.

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	v := tests.SomeValue2 // TestRealmObject2 (value) — sticky N_Readonly propagates
	v.ModifyVal()         // value-receiver: mutates only its local copy
	println("after ModifyVal; v.Field =", v.Field)
}

// Output:
// after ModifyVal; v.Field = init
