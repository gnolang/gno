// PKGPATH: gno.land/r/zrealm_handle_foreign
package zrealm_handle_foreign

// Another realm's non-crossing function taken as a value runs read-only for
// that realm, called here or handed to that realm's gate.

import (
	"chain/runtime"

	"gno.land/r/tests/vm/cbvictim"
)

func try(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return nil
}

func main(cur realm) {
	f := cbvictim.SetOwner
	println("pkgpath:", runtime.FuncPkgPath(f))
	println("call:", try(func() { f("caller") }))
	println("gate:", revive(func() { cbvictim.RunS(cross(cur), f) }))
	println("gate direct:", revive(func() { cbvictim.RunS(cross(cur), cbvictim.SetOwner) }))
	println("owner:", cbvictim.Owner)
}

// Output:
// pkgpath: gno.land/r/tests/vm/cbvictim
// call: gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
// gate: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
// gate direct: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
// owner: victim
