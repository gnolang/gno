// PKGPATH: gno.land/r/zrealm_handle_lib_gate
package zrealm_handle_lib_gate

// A library function taken as a value in the victim's gate acts like the
// same function called by name there: it writes the victim's object.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.LibGate(cross(cur))
	println("lib:", cbvictim.Lib.Field)
}

// Output:
// lib: pwnd-same-pkg
