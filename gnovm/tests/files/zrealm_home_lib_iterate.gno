// PKGPATH: gno.land/r/zrealm_home_lib_iterate
package zrealm_home_lib_iterate

// cbvictim's gate handing its own value to a library that calls it writes:
// the library runs under cbvictim's hold.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunLibIterate(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: kick
