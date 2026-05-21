// PKGPATH: gno.land/r/v3aprobehost
package v3aprobehost

import (
	"gno.land/r/tests/vm/v3aprobe"
)

// v3a security claim, isolated to identity reflection.
//
// /r/v3aprobehost (this realm) calls /r/tests/vm/v3aprobe.Probe()
// without `cross` ceremony. Under v3a's auto-cross walk:
//   - Inside Probe, runtime.Caller() resolves to /r/v3aprobehost
//     (this realm), not the EOA. A WithCross-only walk would have
//     skipped past the /r/→/r/ frame and returned the EOA.
//   - runtime.Self() resolves to /r/tests/vm/v3aprobe (where the
//     code runs), matching m.Realm after Layer-1 borrow.

func main() {
	caller, self := v3aprobe.Probe()

	println("caller.PkgPath:", caller.PkgPath())
	println("self.PkgPath:  ", self.PkgPath())

	// Negative check: the caller is NOT the EOA. If auto-cross
	// detection were absent, the walk would fall through to
	// OriginCaller and Caller.PkgPath() would be empty.
	println("caller.IsUser:", caller.IsUser())
}

// Output:
// caller.PkgPath: gno.land/r/v3aprobehost
// self.PkgPath:   gno.land/r/tests/vm/v3aprobe
// caller.IsUser: false
