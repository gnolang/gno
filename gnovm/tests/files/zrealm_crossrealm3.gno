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
	// NOTE: Modifying an external-realm object via a non-crossing method is
	// allowed by soft crossing.
	somevalue.Modify()
	println(somevalue)
}

// Output:
// (struct{("test_modified" string)} gno.land/r/tests/vm.TestRealmObject)
