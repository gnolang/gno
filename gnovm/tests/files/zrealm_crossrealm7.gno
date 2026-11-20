// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests/v0"
)

func main(cur realm) {
	tests.ModifyTestRealmObject2a()
	println("done")
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): somevalue1<~VPBlock(3,3)>.Field
