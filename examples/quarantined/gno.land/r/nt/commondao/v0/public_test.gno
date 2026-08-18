package commondao

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
)

// TestAssertRunningPath pins the deploy-time guard on the package path.
// Treasury addresses derive from the pkgPath const while Execute mints
// cur.Sub under the path the realm is actually running at, so a copy
// deployed elsewhere would read balances at one address family and send
// from another. The guard turns that into a genesis abort; nothing else
// in the suite notices if it stops doing so.
func TestAssertRunningPath(cur realm, t *testing.T) {
	assertRunningPath() // running at pkgPath: must not panic

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v1"))
	urequire.PanicsWithMessage(t, cur,
		"commondao: realm deployed at gno.land/r/nt/commondao/v1, expected gno.land/r/nt/commondao/v0",
		func() { assertRunningPath() })
}
