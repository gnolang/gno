// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	tests.ModifyTestRealmObject2a()
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): somevalue1<~VPBlock(3,3)>.Field
