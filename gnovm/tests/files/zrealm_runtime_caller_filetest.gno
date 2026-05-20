// PKGPATH: gno.land/r/v3atest
package v3atest

import "chain/runtime"

// Phase A.1 of v3a: verify the new identity-query machine ops behave as
// documented. Caller/Self/Origin are v3a-aligned names that route through
// the existing v2 walk machinery — under the current substrate they
// alias to PreviousRealm/CurrentRealm/OriginCaller-as-Realm, so the
// identity equivalences below should hold.
//
// The point of having Caller/Self/Origin (vs continuing to use
// PreviousRealm/CurrentRealm directly) is the v3a-style code pattern:
// library helpers query runtime.Caller() rather than taking a forgeable
// `rlm realm` parameter. Under the current substrate the underlying
// implementation is the same; the API rename is the safety surface.

func main() {
	// Caller mirrors PreviousRealm.
	caller := runtime.Caller()
	prev := runtime.PreviousRealm()
	println("caller==prev addr:", caller.Address() == prev.Address())
	println("caller==prev path:", caller.PkgPath() == prev.PkgPath())

	// Self mirrors CurrentRealm.
	self := runtime.Self()
	curr := runtime.CurrentRealm()
	println("self==curr addr:", self.Address() == curr.Address())
	println("self==curr path:", self.PkgPath() == curr.PkgPath())

	// Self should report the running realm.
	println("self path:", self.PkgPath())

	// Origin should be the EOA — empty pkgPath.
	origin := runtime.Origin()
	println("origin path empty:", origin.PkgPath() == "")
	println("origin is user:", origin.IsUser())
}

// Output:
// caller==prev addr: true
// caller==prev path: true
// self==curr addr: true
// self==curr path: true
// self path: gno.land/r/v3atest
// origin path empty: true
// origin is user: true
