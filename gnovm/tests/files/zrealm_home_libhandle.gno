// PKGPATH: gno.land/r/zrealm_home_libhandle
package zrealm_home_libhandle

// cbvictim.LibFV is a library handle cbvictim made at init; handed back to
// its gate it has no maker, so the library code runs view-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunLib(cross(cur), cbvictim.LibFV)
	println("lib:", cbvictim.Lib.Field)
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,0)>.Field
