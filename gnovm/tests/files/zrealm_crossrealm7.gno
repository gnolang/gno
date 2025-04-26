// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

func main() {
	crossing()

	tests.ModifyTestRealmObject2a()
	println("done")
}

// Error:
// cannot modify external-realm or non-realm object
