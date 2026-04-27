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
	// NOTE: Modifying an external-realm object via a non-crossing method is
	// allowed by soft crossing.
	somevalue.Modify()
	println(somevalue)
}

// Output:
// readonly(&(struct{("_modified" string)} gno.land/r/tests/vm.TestRealmObject))
