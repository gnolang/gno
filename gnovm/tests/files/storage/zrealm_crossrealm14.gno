// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

func main() {
	// Make1 (entered without crossing) builds a container owned by the
	// crossrealm realm; the /p/ method Touch called from here cannot write it.
	// Nothing is stored or finalized, so no realm's storage changes.
	crossrealm.Make1().Touch().Print()
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): c<VPBlock(1,0)>.A
