// PKGPATH: gno.land/r/borrow
package borrow

import (
	"gno.land/r/tests/vm/crossrealm"
)

func main() {
	crossrealm.UtilityFunc() // not crossing
}

// Error:
// cannot modify external-realm or non-realm object
