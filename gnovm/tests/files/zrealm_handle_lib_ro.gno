// PKGPATH: gno.land/r/zrealm_handle_lib_ro
package zrealm_handle_lib_ro

// A library function taken as a value in the victim's read-only code has no
// maker: the victim's gate runs it view-only and it writes nothing.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunLib(cross(cur), cbvictim.GetLib())
	println("lib:", cbvictim.Lib.Field)
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,0)>.Field
