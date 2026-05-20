// Package v3aclhelper is a regression-test fixture for the v3a-aligned
// ACL pattern: a /p/ helper that identifies its caller via the
// runtime.Caller() machine op instead of accepting a `rlm realm`
// parameter.
//
// This is the omarsy-free shape — there is no `realm`-typed parameter
// for attackers to supply a forged FakeRealm value. The caller's
// identity is queried by the VM at call time and returned as a
// concrete runtime.Realm struct; user code cannot intercept or
// substitute it.
//
// Compare to v2's pattern:
//
//	func (o *Ownable) TransferOwnership(_ int, rlm realm, ...) {
//	    if !rlm.IsCurrent() { ... }
//	    caller := rlm.Previous().Address()
//	    ...
//	}
//
// v3a equivalent (no rlm param, no IsCurrent check):
//
//	func (o *Ownable) TransferOwnership(...) {
//	    caller := runtime.Caller().Address()
//	    ...
//	}
package v3aclhelper

import "chain/runtime"

// CheckCaller is a minimal ACL predicate: returns true iff the
// invoking realm's address matches `expected`. It demonstrates the
// v3a pattern — no `rlm realm` parameter, identity sourced from
// runtime.Caller().
func CheckCaller(expected address) bool {
	return runtime.Caller().Address() == expected
}

// CallerPath returns the calling realm's pkgPath, for tests that want
// to assert which realm crossed in.
func CallerPath() string {
	return runtime.Caller().PkgPath()
}

// CallerIsUser returns true iff the caller chain bottoms out at an
// EOA (i.e., runtime.Caller() has empty pkgPath).
func CallerIsUser() bool {
	return runtime.Caller().IsUser()
}
