// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/p/demo/tests/p_crossrealm"

func main() {
	p_crossrealm.Modify()
}

// Error:
// cannot modify external-realm or non-realm object
