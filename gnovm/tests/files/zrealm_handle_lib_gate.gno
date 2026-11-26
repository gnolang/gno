// PKGPATH: gno.land/r/zrealm_handle_lib_gate
package zrealm_handle_lib_gate

// A library function in a reassigned local in the victim's gate is a value, not
// a call by name: it runs read-only and cannot write the victim's object.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.LibGate(cross(cur))
	println("lib:", cbvictim.Lib.Field)
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,0)>.Field
