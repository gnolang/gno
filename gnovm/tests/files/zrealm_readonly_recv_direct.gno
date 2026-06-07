// PKGPATH: gno.land/r/test/readonlyrecvdirect
package readonlyrecvdirect

// Test A: confirm a pointer obtained from an external /p/ package
// global carries the sticky N_Readonly bit (path 2 of IsReadonlyBy),
// and that a direct field write through it triggers the taint panic
// at PopAsPointer.

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	p := tests.SomeValue3 // *TestRealmObject2 — sticky N_Readonly propagates
	p.Field = "modified"  // expect: readonly taint panic
	println("should not print")
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): p<VPBlock(1,1)>.Field
