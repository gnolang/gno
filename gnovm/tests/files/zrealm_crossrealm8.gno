// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

func main() {
	crossing()

	tests.ModifyTestRealmObject2b()
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): SomeValue2<~VPBlock(3,4)>.Field
