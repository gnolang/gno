// Package v3achecker is a /p/ helper that identifies the caller via
// runtime.Caller() — no realm-typed parameter, no cur/rlm threading.
//
// This is the v3a shape that replaces v2's (_ int, rlm realm, ...)
// pattern. In v1/v2, a /p/ helper using PreviousRealm() or requiring
// cur/rlm threading would misidentify the caller; v3a's Caller()
// correctly resolves the immediate realm that crossed in.
package v3achecker

import "chain/runtime"

// AssertCaller panics if the calling realm's address does not equal
// expected. No cur/rlm parameter — identity is queried from the VM.
func AssertCaller(expected address) {
	if runtime.Caller().Address() != expected {
		panic("unauthorized: " + string(runtime.Caller().Address()))
	}
}

// CallerAddr returns the calling realm's address for test inspection.
func CallerAddr() address {
	return runtime.Caller().Address()
}
