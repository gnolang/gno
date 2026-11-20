// PKGPATH: gno.land/r/zrealm_home_known_ro_relay
package zrealm_home_known_ro_relay

// A reference to the realm's own stored functions: the realm's non-crossing
// helper relays one to its own gate, and S3 refuses the write.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RelayFrom(cbvictim.Table, cur)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
