// PKGPATH: gno.land/r/zrealm_home_ro_relay
package zrealm_home_ro_relay

// cbvictim's read-only code fetches its own stored value; this realm hands it
// to cbvictim's gate by value, so it runs read-only and the write is refused.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	f := cbvictim.GetFV()
	g := f
	cbvictim.RunS(cross(cur), g)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
