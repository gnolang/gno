// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// NOTE: it is valid to persist a pointer to an external object
var somevalue *tests.TestRealmObject

func init() {
	somevalue = &tests.TestRealmObjectValue
}

func main(cur realm) {
	// Modify is /r/tests/vm code entered without crossing, so it runs
	// read-only and its write to the stored object is refused.
	println("modify:", revive(func() { somevalue.Modify() }))
	println(somevalue)
}

// Output:
// modify: gno.land/r/tests/vm's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Field
// &(struct{( string)} gno.land/r/tests/vm.TestRealmObject)
