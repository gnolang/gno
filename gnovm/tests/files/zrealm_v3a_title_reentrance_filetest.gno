// PKGPATH: gno.land/r/v3aattacker
package v3aattacker

import (
	"gno.land/r/tests/vm/v3avictim"
)

// v3a Title()-attack closure test.
//
// This reproduces the original Title() + PreviousRealm() issue: a /p/
// helper (v3achecker.AssertCaller) identifies the caller via a stack
// walk — no cur/rlm parameter. In v1/v2, a /p/ helper using
// PreviousRealm() or the forged-value cur/rlm workaround would
// misidentify the caller; v3a's Caller() correctly resolves the
// immediate realm that crossed in.
//
// Scenario:
//   1. r/v3avictim delegates auth to /p/v3achecker.AssertCaller,
//      which uses runtime.Caller() without any realm parameter.
//   2. r/v3aattacker (this realm) registers a callback that re-enters
//      v3avictim.SetSecret.
//   3. We trigger RunCallback. The victim invokes our callback.
//      Our callback re-enters SetSecret → AssertCaller.
//
// Chain: r/v3aattacker → r/v3avictim → /p/v3achecker.AssertCaller
//
// Inside AssertCaller, runtime.Caller() returns r/v3aattacker —
// the realm that crossed into the victim — not the EOA.
// The auth check panics: attacker ≠ owner.
//
// In v1, PreviousRealm() inside the /p/ helper would skip non-crossing
// frames and return the EOA. In v2, the workaround was (_ int, rlm realm, ...)
// — passing identity as a forgeable value. v3a eliminates both.

func main() {
	// Plant the malicious callback.
	v3avictim.SetCallback(func(newOwner address) {
		// We are running in r/v3aattacker's realm (Layer 1 borrow).
		// Re-enter the victim's guarded function.
		v3avictim.SetSecret("pwnd")
		// Inside SetSecret → /p/v3achecker.AssertCaller(owner):
		//   runtime.Caller() = r/v3aattacker ≠ owner → PANIC
	})

	// Trigger — the victim invokes our callback.
	r := revive(func() { v3avictim.RunCallback(address("bob")) })
	println("re-entrance blocked:", r != nil)
	println("panic msg:", r) // shows "unauthorized: <attacker-addr>"

	// Secret must still be the original — the attack was blocked.
	println("secret unchanged:", v3avictim.ReadSecret() == "initial-secret")
}

// Output:
// re-entrance blocked: true
// panic msg: unauthorized: g1hxu9u0uuw3vuedlzxv455cp9egkagnwpwvh8dx
// secret unchanged: true
