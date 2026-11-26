package commondao

import (
	"chain"
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

// TestAssertExecutor pins the guard every crossing write helper opens with.
// A helper's panic leaves a cross call, so a test cannot recover it; the
// guard is tested here directly instead.
func TestAssertExecutor(cur realm, t *testing.T) {
	urequire.NotPanics(t, cur, func() { assertExecutor(daoAddress(301), 301) })
	urequire.PanicsWithMessage(t, cur, "commondao: caller is not the DAO's executor",
		func() { assertExecutor(daoAddress(302), 301) })
	urequire.PanicsWithMessage(t, cur, "commondao: caller is not the DAO's executor",
		func() { assertExecutor(chain.PackageAddress(pkgPath), 301) })
}

// TestAssertSub pins the check executors run on sub before crossing with it.
func TestAssertSub(cur realm, t *testing.T) {
	sub := cur.Sub(subpathOf(301))
	urequire.NotPanics(t, cur, func() { assertSub(0, sub) })
	urequire.PanicsWithMessage(t, cur, "commondao: executor sub is not current",
		func() { assertSub(0, nil) })
	stale := mintSub(cross(cur), 301)
	urequire.PanicsWithMessage(t, cur, "commondao: executor sub is not current",
		func() { assertSub(0, stale) })
}
