// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests"
)

// NOTE: it is valid to persist external realm types.
var somevalue *tests.TestRealmObject

func init() {
	somevalue = &tests.TestRealmObjectValue
}

func main() {
	// NOTE: but it is invalid to modify it directly.
	somevalue.Field = "test"
	println(somevalue)
}

// Error:
// cannot modify external-realm or non-realm object
