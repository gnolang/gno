// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	crossrealm_b.IncrGlobal()
	println("should not print")
}

// Output:

// Error:
// cannot modify external-realm or non-realm object
