// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// A callee cannot construct a spend-capable banker over its caller:
// NewBanker requires the realm value to be the live cur, and
// cur.Previous() (the caller) is not IsCurrent. Without this gate a
// called realm could drain whoever called it.
func main(cur realm) {
	crossrealm_b.TryBankerOnPrevious(cross(cur))
}

// Error:
// banker can only be instantiated for the current realm
