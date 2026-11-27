// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Negative control: without recover(), the refused non-crossing write
// (crossrealm_d entered without crossing is read-only) aborts everything.

import (
	"gno.land/r/tests/vm/crossrealm_d"
)

func main(cur realm) {
	println("counter:", crossrealm_d.GetCounter(cross(cur)))

	// No recover — the cross-realm panic aborts everything.
	crossrealm_d.SetCounter(0)

	// This line is never reached.
	println("counter after:", crossrealm_d.GetCounter(cross(cur)))
}

// Output:
// counter: 100

// Error:
// gno.land/r/tests/vm/crossrealm_d's code was entered without crossing and is read-only (call its crossing function): counter<~VPBlock(3,0)>
