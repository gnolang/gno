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
	// NOTE: it is invalid to modify it even using the external realm function,
	// because the value was retrieved by a selector. XXX reconsider
	somevalue.Modify()
	println(somevalue)
}

// OutputXXX:
// &(struct{("_modified" string)} gno.land/r/demo/tests.TestRealmObject)

// Error:
// cannot modify external-realm or non-realm object
