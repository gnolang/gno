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
	// NOTE: it is to modify external realm resident object using its method.
	somevalue.Modify()
	println(somevalue)
}

// Output:
// &(struct{("_modified" string)} gno.land/r/demo/tests.TestRealmObject)
