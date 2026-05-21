// PKGPATH: gno.land/r/v3acallera
package v3acallera

import (
	"gno.land/r/tests/vm/v3acallee"
)

// v3a auto-cross identity test.
//
// r/v3acallera calls r/v3acallee's exported functions with plain
// arguments — no cross keyword, no cur realm parameter, no rlm realm
// parameter. Under v3a, the preprocessor auto-crosses on every direct
// foreign /r/ call, so the callee's runtime.Caller() correctly resolves
// to r/v3acallera and runtime.Self() resolves to r/v3acallee.
//
// Under v1, the same calls would shift m.Realm but PreviousRealm() would
// not reflect the shift for non-crossing calls. Under v2, the callee
// would require explicit cross(fn)(args) or cross2(rlm) ceremony to
// correctly surface the caller's identity.
//
// This test demonstrates that v3a eliminates the ceremony while
// preserving identity correctness.

func main() {
	// Identity reflection: Caller and Self without any cross keyword.
	println("callee.Self:   ", v3acallee.WhoAmI())
	println("callee.Caller: ", v3acallee.WhoCalledMe())

	// Configure the callee: only r/v3acallera may call SetValue.
	v3acallee.SetOwner("gno.land/r/v3acallera")

	// Authorized: caller matches configured owner.
	v3acallee.SetValue("hello-v3a")
	println("after auth set:", v3acallee.GetValue() == "hello-v3a")

	// Change the owner — now r/v3acallera is no longer authorized.
	v3acallee.SetOwner("gno.land/r/someoneelse")

	// Unauthorized call should panic. Use revive to catch the
	// cross-realm abort.
	r := revive(func() { v3acallee.SetValue("should-fail") })
	println("unauthorized blocked:", r != nil)

	// Value unchanged by the rejected call.
	println("value unchanged:", v3acallee.GetValue() == "hello-v3a")
}

// Output:
// callee.Self:    gno.land/r/tests/vm/v3acallee
// callee.Caller:  gno.land/r/v3acallera
// after auth set: true
// unauthorized blocked: true
// value unchanged: true
