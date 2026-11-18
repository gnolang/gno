// PKGPATH: gno.land/r/zrealm_handle_mexpr
package zrealm_handle_mexpr

// A method expression taken as a value in the victim's read-only code writes
// nothing in its gate; one taken in a gate writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	println("read-only:", revive(func() { cbvictim.RunM(cross(cur), cbvictim.GetMExpr()) }))
	println("name: [" + cbvictim.Box.Name + "]")
	println("gate:", revive(func() { cbvictim.RunM(cross(cur), cbvictim.GateMExpr(cross(cur))) }))
	println("name: [" + cbvictim.Box.Name + "]")
}

// Output:
// read-only: gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
// name: []
// gate: undefined
// name: [picked]
