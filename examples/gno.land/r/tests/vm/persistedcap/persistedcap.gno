// Package persistedcap stores a closure (constructed by a /p/ factory)
// at init, persisting the closure's captured HIV as part of this
// realm's state. Foreign-realm callers can grab the closure via
// GetCounter() and invoke it; the closure body's write to its
// captured HIV is REJECTED by the readonly check because the HIV is
// real (NewTime>0) and belongs to a different realm than the caller's
// m.Realm. The unreal-HIV exception only applies while the HIV is
// transient — persistence elevates it to a normal realm-owned slot.
package persistedcap

import "gno.land/p/demo/tests/p_closurecap"

var counter func() int

func init() {
	counter = p_closurecap.MakeCounter(0)
}

// GetCounter returns the persisted closure. The caller invokes it
// directly — because the closure's FuncLit lives in /p/ p_closurecap,
// PushFrameCall's borrow rule does NOT switch m.Realm to persistedcap
// at invocation. The closure body's write to the captured HIV (PkgID
// = persistedcap, persisted) therefore runs under the caller's
// m.Realm, which is the persisted-closure-capture cross-realm case.
//
// SECURITY: deliberate VM-parity-test infrastructure — do NOT copy this
// pattern in production code. See gno.land/r/tests/vm/crossrealm
// Closure for the broader rationale.
func GetCounter() func() int {
	return counter
}
