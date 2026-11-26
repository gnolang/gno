// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// NOTE: it is valid to persist external realm types.
var somevalue *tests.TestRealmObject

func init() {
	somevalue = &tests.TestRealmObjectValue
}

func main(cur realm) {
	// NOTE: but it is invalid to modify it directly.
	somevalue.Field = "test"
	println(somevalue)
}

// Error:
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ or standard-library package's object cannot be written after its init): somevalue<~VPBlock(3,0)>.Field
