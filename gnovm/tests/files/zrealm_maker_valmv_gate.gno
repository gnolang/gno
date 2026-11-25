// PKGPATH: gno.land/r/zrealm_maker_valmv_gate
package zrealm_maker_valmv_gate

// A value-receiver method value bound inside a crossing function has the
// victim as binder; handed back by another realm it has none: read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GateValMV(cross(cur)))
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.P.Name
