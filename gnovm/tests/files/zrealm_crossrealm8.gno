// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/p/demo/tests/v0"
)

func main(cur realm) {
	tests.ModifyTestRealmObject2b()
	println("done")
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): SomeValue2<~VPBlock(3,4)>.Field
