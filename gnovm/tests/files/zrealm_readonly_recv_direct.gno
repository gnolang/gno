// PKGPATH: gno.land/r/test/readonlyrecvdirect
package readonlyrecvdirect

// Test A: confirm a pointer obtained from an external /p/ package
// global carries the sticky N_Readonly bit (path 2 of IsReadonlyBy),
// and that a direct field write through it triggers the taint panic
// at PopAsPointer.

import (
	"gno.land/p/demo/tests/v0"
)

func main(cur realm) {
	p := tests.SomeValue3 // *TestRealmObject2 — sticky N_Readonly propagates
	p.Field = "modified"  // expect: readonly taint panic
	println("should not print")
}

// Error:
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ package's object cannot be written after its init): p<VPBlock(1,1)>.Field
