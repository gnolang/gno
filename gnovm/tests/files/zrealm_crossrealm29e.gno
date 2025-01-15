// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var b = 1

type Local_S struct {
	name string
}

func main() {
	f := func() bool {
		c := b
		var ls Local_S
		return true
	}
	crossrealm.SetCallback(f)
}

// Error:
// cannot attach function contains object from external realm
