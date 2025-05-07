// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

func main() {
	crossing()

	tests.ModifyTestRealmObject2c()
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): SomeValue3<~VPBlock(3,5)>.Field
