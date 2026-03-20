// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Negative control: without recover(), the cross-realm readonly
// check properly aborts the entire transaction.

import (
	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	result := crossrealm_d.GetResult(cross)
	println("before:", result)

	// No recover — the cross-realm panic aborts everything.
	crossrealm_d.SetVotes(99, 1)

	// This line is never reached.
	println("after:", crossrealm_d.GetResult(cross))
}

// Output:
// before: rejected

// Error:
// cannot directly modify readonly tainted object (w/o method): yesVotes<~VPBlock(3,0)>
