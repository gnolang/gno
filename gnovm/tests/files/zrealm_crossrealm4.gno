// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests"
)

// NOTE: it is valid to persist a pointer to an external object
var somevalue *tests.TestRealmObject

func init() {
	somevalue = &tests.TestRealmObjectValue
}

func main() {
	crossing()

	// NOTE: it is valid to modify it using the external realm function.
	somevalue.Modify()
	println(somevalue)
}

// Output:
// readonly(&(struct{("_modified" string)} gno.land/r/demo/tests.TestRealmObject))
