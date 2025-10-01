// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// NOTE: it is valid to persist external realm types.
var somevalue tests.TestRealmObject

func main(cur realm) {
	somevalue.Field = "test"
	println(somevalue)
}

// Output:
// (struct{("test" string)} gno.land/r/tests/vm.TestRealmObject)
