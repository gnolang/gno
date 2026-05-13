// Package curchain_b is the terminal hop in the multi-level cross-call
// chain test helpers. It prints the full captured chain visible from this
// realm, and asserts parity with runtime.CurrentRealm() /
// runtime.PreviousRealm() at the cross-callee position.
package curchain_b

import "chain/runtime"

// B is a crossing function that prints the captured chain reached via
// .Previous() walks and asserts parity with runtime.CurrentRealm() /
// runtime.PreviousRealm() at the cross-callee position. The chain-root
// case is covered separately by zrealm_cur_parity.
//
// Chain layout for callsite main -> curchain_a.A -> curchain_b.B:
//   cur                  = B
//   cur.Previous()       = A
//   cur.Previous()^2     = main
//   cur.Previous()^3     = EOA origin (PkgPath() == "")
//   cur.Previous()^4     = panics (origin has no further previous)
func B(cur realm) {
	println("B cur.PkgPath:", cur.PkgPath())
	p1 := cur.Previous()
	println("B Previous().PkgPath:", p1.PkgPath())
	p2 := p1.Previous()
	println("B Previous().Previous().PkgPath:", p2.PkgPath())
	p3 := p2.Previous()
	println("B Previous()^3 PkgPath empty (EOA origin):", p3.PkgPath() == "")
	func() {
		defer func() {
			r := recover()
			println("B Previous()^4 panics past origin:", r != nil)
		}()
		_ = p3.Previous()
	}()

	// Parity at cross-callee position: cur agrees with
	// runtime.CurrentRealm(), and cur.Previous() (the caller realm
	// A) agrees with runtime.PreviousRealm().
	rc := runtime.CurrentRealm()
	rp := runtime.PreviousRealm()
	println("B parity cur==Current:",
		string(cur.Address()) == string(rc.Address()) &&
			cur.PkgPath() == rc.PkgPath())
	println("B parity prev==Previous:",
		string(p1.Address()) == string(rp.Address()) &&
			p1.PkgPath() == rp.PkgPath())
}
