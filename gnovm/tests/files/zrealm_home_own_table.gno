// PKGPATH: gno.land/r/zrealm_home_own_table
package zrealm_home_own_table

// cbvictim's gate running its own table writes: the values never left.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunOwnTable(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: own-table
