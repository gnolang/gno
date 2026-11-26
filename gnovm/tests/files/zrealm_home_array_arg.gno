// PKGPATH: gno.land/r/zrealm_home_array_arg
package zrealm_home_array_arg

// An array passed by value is copied under the caller's hold, element by
// element.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunArr(cross(cur), cbvictim.Arr)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
