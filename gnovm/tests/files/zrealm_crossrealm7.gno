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
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): somevalue1<~VPBlock(3,3)>.Field
