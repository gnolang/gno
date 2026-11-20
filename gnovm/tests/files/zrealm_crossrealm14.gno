// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

func main(cur realm) {
	// Make1 builds the container in the crossrealm realm and returns it.
	// It is that realm's object even though nothing stores it, so the /p/
	// method Touch called from here cannot write it (explicit-entry rule).
	crossrealm.Make1().Touch().Print()
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): c<VPBlock(1,0)>.A
