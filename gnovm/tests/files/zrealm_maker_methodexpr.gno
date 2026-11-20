// PKGPATH: gno.land/r/zrealm_maker_methodexpr
package zrealm_maker_methodexpr

// Another realm's method expression used as a value runs read-only for it.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	f := (*cbvictim.T).SetName
	f(cbvictim.Box, "picked")
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
