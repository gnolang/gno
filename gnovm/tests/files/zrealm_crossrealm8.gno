// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	tests.ModifyTestRealmObject2b()
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): SomeValue2<~VPBlock(3,4)>.Field
