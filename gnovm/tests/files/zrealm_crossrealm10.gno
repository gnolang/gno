// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	tests.ModifyTestRealmObject2c()
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): SomeValue3<~VPBlock(3,5)>.Field
