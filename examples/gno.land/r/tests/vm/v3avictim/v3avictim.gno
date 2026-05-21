// Package v3avictim is a victim realm for the v3a Title()-attack
// filetest. It holds guarded state (secret) and exposes a callback
// slot (onTransfer) that an attacker can register — the injection point.
//
// Auth is delegated to /p/demo/tests/v3achecker.AssertCaller, a /p/
// helper that uses runtime.Caller() without any cur/rlm parameter.
// This reproduces the original Title() shape: a /p/ helper identifying
// the caller via a stack walk. v1's PreviousRealm() and v2's
// unsafe.PreviousRealm() would skip the /p/ frame and misidentify the
// caller; v3a's Caller() walk correctly returns the immediate realm
// that crossed in.
//
// The attacker registers a callback that re-enters SetSecret. The
// callback runs in r/v3aattacker's realm (Layer 1 borrow), then
// auto-crosses back into r/v3avictim. Inside SetSecret → AssertCaller,
// runtime.Caller() returns r/v3aattacker, not the EOA, so the check
// panics and the attack is blocked.
package v3avictim

import (
	"gno.land/p/demo/tests/v3achecker"
)

// The designated owner — only this address can set the secret.
var owner = address("g1deployer000000000000000000000000000000")

// The guarded state.
var secret string

// The callback slot — an attacker registers here.
var onTransfer func(address)

func init() {
	secret = "initial-secret"
}

// SetCallback lets a foreign realm register a callback.
func SetCallback(fn func(address)) {
	onTransfer = fn
}

// RunCallback invokes the registered callback. Intentionally unguarded —
// simulates a function that legitimately invokes caller-supplied hooks
// (e.g. a banker send, a notification dispatch).
func RunCallback(newOwner address) {
	if onTransfer != nil {
		onTransfer(newOwner)
	}
}

// SetSecret is the guarded function. Auth is delegated to a /p/ helper
// (v3achecker.AssertCaller) that uses runtime.Caller() — no cur/rlm
// parameter. Under v3a, Caller() correctly resolves to the immediate
// realm that crossed in. Under v1, PreviousRealm() would walk past the
// /p/ frame and misidentify the caller.
func SetSecret(newSecret string) {
	v3achecker.AssertCaller(owner)
	secret = newSecret
}

// ReadSecret returns the current secret for test verification.
func ReadSecret() string {
	return secret
}
