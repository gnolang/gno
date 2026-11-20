// PKGPATH: gno.land/r/zrealm_home_known_relay_table
package zrealm_home_known_relay_table

// Known gap: a reference to the realm's own stored functions is not covered.
// cbother hands cbvictim's table back by reference; S3 refuses the write.

import (
	"gno.land/r/tests/vm/cbother"
	"gno.land/r/tests/vm/cbvictim"
)

func main(cur realm) {
	cbother.RelayTable(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
