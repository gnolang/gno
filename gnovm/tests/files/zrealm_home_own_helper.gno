// PKGPATH: gno.land/r/zrealm_home_own_helper
package zrealm_home_own_helper

// cbvictim's gate handing its own value to its own helper writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunOwnHelper(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: own-helper
