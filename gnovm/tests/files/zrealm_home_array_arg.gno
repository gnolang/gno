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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
