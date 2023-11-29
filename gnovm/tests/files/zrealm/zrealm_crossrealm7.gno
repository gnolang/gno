// PKGPATH: gno.land/r/crossrealm_test
package main

import (
	"gno.land/p/demo/tests"
)

func main() {
	tests.ModifyTestRealmObject2a()
	println("done")
}

// Error:
// cannot modify external-realm or non-realm object
