// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// Pointer-typed foreign var initialized via a foreign constructor that
// returns &CompositeLit: the object is a fresh heap object owned by
// /r/tests/vm, not a slot in its package block (contrast crossrealm4).
// Constructing it without crossing is allowed; modifying it later via the
// non-crossing method Modify is refused, because /r/tests/vm's code was
// entered without crossing and is read-only.
var somevalue *tests.TestRealmObject

func init() {
	somevalue = tests.NewTestRealmObject()
}

func main(cur realm) {
	println("modify:", revive(func() { somevalue.Modify() }))
	println(somevalue)
}

// Output:
// modify: gno.land/r/tests/vm's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Field
// &(struct{("initial" string)} gno.land/r/tests/vm.TestRealmObject)
