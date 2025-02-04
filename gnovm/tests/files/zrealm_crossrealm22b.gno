// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var b = 1

type Local_S struct {
	name string
}

func main() {
	var ls Local_S
	f := func() bool {
		c := b
		return true
	}
	crossrealm.SetCallback(f)
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm
