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
	// NOTE: but it is invalid to modify it using an external realm function.
	somevalue.Modify()
	println(somevalue)
}

// Error:
// cannot modify external-realm or non-realm object
