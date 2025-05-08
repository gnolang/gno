// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

func main() {
	crossing()

	tests.ModifyTestRealmObject2a()
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): somevalue1<~VPBlock(3,3)>.Field
