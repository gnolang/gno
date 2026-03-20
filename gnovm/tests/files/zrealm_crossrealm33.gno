// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	crossrealm_b.IncrGlobal()
	println("should not print")
}

// Output:

// Error:
// cannot directly modify readonly tainted object (w/o method): n<~VPBlock(3,10)>
