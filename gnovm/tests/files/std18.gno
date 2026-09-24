// PKGPATH: gno.land/r/test
package test

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	rtests "gno.land/r/tests/vm"
)

// testing.SetRealm(NewCodeRealm(p)) below the entry realm simulates a code
// caller p. Naming the entry realm itself is the realm calling its own
// function, still an origin call; a user caller is what main already stands
// in for. Another realm's path is refused: std19. Calls are direct because
// main's own callees (revive, closures) run as r/test and would make r/test
// the entry realm.
func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/tests/vm"))
	rtests.CallAssertOriginCall(cross(cur))
	println("self code realm ok")

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("user")))
	rtests.CallAssertOriginCall(cross(cur))
	println("user ok")
}

// Output:
// self code realm ok
// user ok
