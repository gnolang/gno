// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var b = 1

type Local_S struct {
	name string
}

var f = func() bool {
	c := b
	var ls Local_S
	return true
}

func main() {
	crossrealm.SetCallback(f)
	println("ok")
}

// Output:
// ok

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/func"]
// switchrealm["gno.land/r/crossrealm_test"]
