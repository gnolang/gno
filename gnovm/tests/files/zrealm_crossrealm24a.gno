// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

var b0 *crossrealm.Bar = &crossrealm.Bar{A: 11}

func init() {
}

func main() {
	print(".")
}

// Error:
// cannot attach a reference to an unreal object from an external realm
