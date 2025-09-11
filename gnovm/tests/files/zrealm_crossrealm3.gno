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
	// NOTE: it is also valid to modify it using an external realm function.
	somevalue.Modify()
	println(somevalue)
}

// Output:
// (struct{("test_modified" string)} gno.land/r/tests/vm.TestRealmObject)
