// PKGPATH: gno.land/r/zrealm_maker_valmv_gate
package zrealm_maker_valmv_gate

// A value-receiver method value bound inside a crossing function writes for
// the victim's realm.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GateValMV(cross(cur)))
	println("name:", cbvictim.Box.Name)
}

// Output:
// name: picked
