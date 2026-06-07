// PKGPATH: gno.land/r/test/readonlyrecvmethod
package readonlyrecvmethod

// Method-dispatch mutation on a real /p/-stamped receiver is rejected
// post-init by the DidUpdate /p/-immutability gate.
//
// Before the gate, p.Modify() laundered the sticky N_Readonly bit
// (the bit lives on the caller-side TV, not on the receiver-as-args[0]
// inside the callee's frame) and the in-method `o2.Field = "modified"`
// passed PopAsPointer's check. The mutation succeeded in-tx but never
// persisted — /p/-state reloads fresh on every tx.
//
// With the gate: any DidUpdate firing inside a /p/-frozen package
// (PkgID IsImmutablePkg && !IsStdlibPkg) on a real (NewTime>0) object
// during StageRun panics. Stdlib stays exempt; /r/-stored receivers
// using /p/-helper methods are unaffected (their PkgID is /r/).

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	p := tests.SomeValue3 // *TestRealmObject2 in /p/demo/tests
	p.Modify()            // panics at DidUpdate
	println("unreachable; Field =", p.Field)
}

// Error:
// cannot mutate gno.land/p/demo/tests: package is immutable post-init
