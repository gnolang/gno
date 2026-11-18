// PKGPATH: gno.land/r/zrealm_handle_initvar
package zrealm_handle_initvar

// KNOWN (Q11): a value made at init has the realm as maker; handed back to a
// gate that runs callbacks, it writes. Own function and library variants.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.FV)
	println("owner:", cbvictim.Owner)
	cbvictim.RunLib(cross(cur), cbvictim.LibFV)
	println("lib:", cbvictim.Lib.Field)
}

// Output:
// owner: picked
// lib: pwnd-same-pkg
