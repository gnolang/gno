// PKGPATH: gno.land/r/zrealm_handle_initvar
package zrealm_handle_initvar

// A value made at init has the realm as maker; handed back by another realm to
// a gate that runs callbacks, it has none and writes nothing (read-only).

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.FV)
	println("owner:", cbvictim.Owner)
	cbvictim.RunLib(cross(cur), cbvictim.LibFV)
	println("lib:", cbvictim.Lib.Field)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
