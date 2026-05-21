// Package v3aprobe is a v3a identity-reflection helper. It exports a
// single function that returns (Caller, Self) from inside its body, so
// a /r/A caller can observe what runtime.Caller() / Self() resolve to
// when an /r/A → /r/B exported call auto-crosses under v3a.
//
// The probe is intentionally tiny: no state, no ACL, just the two
// machine ops. This isolates the auto-cross walk's behavior from the
// ownership / authorization pattern exercised by v3acanary.
package v3aprobe

import "chain/runtime"

// Probe returns the realm of the caller and the realm running this
// code, as seen by the v3a primitives:
//   - caller.PkgPath() is the calling /r/ realm (auto-cross detected)
//   - self.PkgPath() is gno.land/r/tests/vm/v3aprobe (this realm)
func Probe() (caller runtime.Realm, self runtime.Realm) {
	return runtime.Caller(), runtime.Self()
}

// DeepProbe returns Caller, Self, and CallerN(2) — the realm two
// transitions up the call chain — observed from inside /r/v3aprobe's
// body. Used by walk-cornerstone tests to assert multi-hop chain
// resolution.
func DeepProbe() (caller runtime.Realm, self runtime.Realm, prev runtime.Realm) {
	return runtime.Caller(), runtime.Self(), runtime.CallerN(2)
}
