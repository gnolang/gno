// PKGPATH: gno.land/r/zrealm_home_known_ro_relay
package zrealm_home_known_ro_relay

// A reference to the realm's own stored functions: the realm's non-crossing
// helper relays one to its own gate, and R2a refuses the write.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RelayFrom(cbvictim.Table, cur)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
