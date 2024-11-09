// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type local struct{}

var a local

var b0 crossrealm.Bar // run declarations

func main() {
	print(".")
}

// Error:
// !!!cross realm while attach object
