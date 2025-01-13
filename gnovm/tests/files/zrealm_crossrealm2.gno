// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests"
)

// NOTE: it is invalid to persist value of external realm types.
var somevalue tests.TestRealmObject

func init() {
}

func main() {
}

// Error:
// cannot attach a value of a type defined by another realm
