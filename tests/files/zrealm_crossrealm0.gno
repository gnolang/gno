// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/tests"
)

// NOTE: it is valid to persist external realm types.
var somevalue tests.TestRealmObject

func main() {
	somevalue.Field = "test"
	println(somevalue)
}

// Output:
// struct{("test" string)}
