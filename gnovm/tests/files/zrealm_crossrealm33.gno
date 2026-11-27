// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// IncrGlobal is entered without crossing, so crossrealm_b runs read-only and
// its write to its own package var is refused.
func main(cur realm) {
	crossrealm_b.IncrGlobal()
	println("should not print")
}

// Error:
// gno.land/r/tests/vm/crossrealm_b's code was entered without crossing and is read-only (call its crossing function): n<~VPBlock(3,10)>
