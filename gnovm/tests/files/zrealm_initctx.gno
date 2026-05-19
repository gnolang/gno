// PKGPATH: gno.land/r/demo/testinit
package testinit

import (
	"chain/runtime"

	tests "gno.land/r/tests/vm"
)

var (
	addr     = address("test")
	addrInit = address("addrInit")
)

func init(cur realm) {
	addr = runtime.OriginCaller()
	addrInit = tests.InitOriginCaller(cross(cur))
}

func main(cur realm) {
	// `addr` is set by init via runtime.OriginCaller() — now reads the
	// DefaultCaller because the filetest framework sets OriginCaller
	// before RunMemPackage (consistent with the package-main path).
	// `addrInit` is set via a cross-call that mints a fresh cur whose
	// authority is the package itself, so the cross-call's view of
	// OriginCaller resolves through the new realm context (empty here).
	println(addr)
	println(addrInit)
	println(".") // trailing newline is needed.
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// .

// Realm:
// finalizerealm["gno.land/r/demo/testinit"]
