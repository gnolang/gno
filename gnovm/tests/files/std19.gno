// PKGPATH: gno.land/r/test
package test

import (
	"testing"

	rtests "gno.land/r/tests/vm"
)

// A code caller from another realm below the entry realm is a realm on the
// stack, as subtests -> r/tests/vm would be on chain. Direct call: see std18.
func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/tests/vm/subtests"))
	rtests.CallAssertOriginCall(cross(cur))
}

// Error:
// invalid non-origin call
