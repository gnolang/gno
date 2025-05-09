// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

func main() {
	crossing()

	tests.SomeValue2.Field = "modified"
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/p/demo/tests) package{})).SomeValue2.Field
