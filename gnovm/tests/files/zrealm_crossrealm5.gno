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
	crossing()

	// NOTE: but it is invalid to modify it directly.
	somevalue.Field = "test"
	println(somevalue)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): somevalue<~VPBlock(3,0)>.Field
