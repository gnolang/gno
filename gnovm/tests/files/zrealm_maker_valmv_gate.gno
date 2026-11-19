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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.P.Name
