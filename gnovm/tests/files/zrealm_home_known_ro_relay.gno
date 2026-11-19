// PKGPATH: gno.land/r/zrealm_home_known_ro_relay
package zrealm_home_known_ro_relay

// Known gap: a reference to the realm's own stored functions is not covered.
// Here the realm's non-crossing helper relays one of them to its own gate.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RelayFrom(cbvictim.Table, cur)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: picked
