// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

var f = func() crossrealm.Fooer {
	return crossrealm_b.Fooer
}

func main() {
	crossrealm.SetFooerGetter(f)
	println("ok")
}

// Error:
// cannot attach objects by value from external realm
