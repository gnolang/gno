// PKGPATH: gno.land/r/test/readonlyrecvmethod
package readonlyrecvmethod

// Test B: same setup as Test A — pointer with sticky N_Readonly from
// an external /p/ package — but call a method that mutates instead
// of writing the field directly. Result: the method DOES mutate.
// The sticky N_Readonly bit on the caller-side TV does not propagate
// onto the receiver-as-args[0] inside the callee's frame, so the
// in-method `o2.Field = "modified"` passes PopAsPointer's check.
// Method dispatch effectively launders the sticky-bit taint —
// readonly-via-N is therefore a direct-access-only check, not a
// transitive mutation guard.

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	p := tests.SomeValue3 // *TestRealmObject2 — sticky N_Readonly propagates
	p.Modify()            // o2.Field = "modified" — does the receiver's bit survive?
	println("after Modify; Field =", p.Field)
}

// Output:
// after Modify; Field = modified
