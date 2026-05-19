// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Negative control: without recover(), the cross-realm readonly
// check properly aborts the entire transaction.

import (
	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	println("counter:", crossrealm_d.GetCounter(cross))

	// No recover — the cross-realm panic aborts everything.
	crossrealm_d.SetCounter(0)

	// This line is never reached.
	println("counter after:", crossrealm_d.GetCounter(cross))
}

// Output:
// counter: 100

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): counter<~VPBlock(3,0)>
