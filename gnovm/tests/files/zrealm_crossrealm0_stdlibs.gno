// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests"
)

// NOTE: it is invalid to persist external realm types.
var somevalue tests.TestRealmObject

func main() {
	somevalue.Field = "test"
	println(somevalue)
}

// Error:
// should not happen while attempting to attach new real object from external realm
