// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests"
)

// NOTE: it is valid to persist external realm types.
var somevalue tests.TestRealmObject

func init() {
	somevalue.Field = "test"
}

func main() {
	crossing()

	// NOTE: but it is invalid to modify it using an external realm function.
	cross(tests.ModifyTestRealmObject)(&somevalue)
	println(somevalue)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): t<VPBlock(1,0)>.Field
