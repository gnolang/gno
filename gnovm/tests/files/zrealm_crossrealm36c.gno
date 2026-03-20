// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Negative control: without recover(), the cross-realm guard
// properly aborts the entire transaction. State is NOT corrupted.
//
// This proves that recover() is the enabling factor — the guard
// itself works, but it fires AFTER the mutation, so recover()
// swallows the panic while the damage is already done.

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
// cannot modify external-realm or non-realm object
