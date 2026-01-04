// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// NOTE: it is valid to persist external realm types.
var somevalue tests.TestRealmObject

func init() {
	somevalue.Field = "test"
}

func main(cur realm) {
	// NOTE: but it is invalid to modify it using an external realm function.
	tests.ModifyTestRealmObject(cross, &somevalue)
	println(somevalue)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): t<VPBlock(1,1)>.Field
