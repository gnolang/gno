// PKGPATH: gno.land/r/launderrh_pdr_recover_retry
package launderrh_pdr_recover_retry

// Vector 1: panic during /r/-victim-borrowed body, recover via
// attacker-defined try/recover, attempt to retry write.
//
// Observation: cross-realm panics (panic that originates inside a
// realm-borrowed frame, then unwinds across the boundary) are NOT
// catchable by a regular defer/recover in the calling realm. They
// hit isRealmBoundary in doOpReturnCallDefers and abort with
// makeUnhandledPanicError. So an attacker's main() cannot catch
// the readonly panic that fires inside ApplyHook's callback frame.
//
// This is a known/intentional invariant: a write attempt across
// realms aborts the transaction, not catchable. Confirms attacker
// cannot use recover() to "swallow" the readonly panic and retry.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	// Try to recover the readonly panic in attacker realm: this
	// SHOULD NOT work — the panic escapes across the realm
	// boundary and the recovery is bypassed.
	defer func() {
		if r := recover(); r != nil {
			println("attacker recovered:", r)
		} else {
			println("no recover")
		}
	}()

	launderrvictim.ApplyHook(func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-callback"
	})

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
