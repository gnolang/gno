// Package v3acomposed is a transparent /p/ wrapper used by the
// walk-cornerstone test. A /p/ function has no realm authority of
// its own; it inherits the calling realm. The test threads a
// function value into Run so the actual /r/-declared call happens
// inside the /p/ frame's body — letting the test assert that an
// intermediate /p/ frame does NOT appear as a realm transition in
// the walk.
//
// /p/ cannot import /r/, so we accept the destination as a function
// value rather than calling /r/v3aprobe directly.
package v3acomposed

import "chain/runtime"

// Run invokes the supplied probe function from inside a /p/ frame.
// No Layer-1 (callee is /p/-declared) and no Layer-2 (no receiver)
// fires for the entry to Run, so m.Realm stays at the caller's
// realm through Run's body. When fn is invoked and fn is an
// /r/-declared function value, Layer-1 fires at that point.
func Run(fn func() (runtime.Realm, runtime.Realm)) (caller runtime.Realm, self runtime.Realm) {
	return fn()
}
