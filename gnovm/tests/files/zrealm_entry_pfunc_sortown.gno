// PKGPATH: gno.land/r/zrealm_entry_pfunc_sortown
package zrealm_entry_pfunc_sortown

// A realm calling a stdlib function through a reassigned local holds a value,
// not a call by name: sort.Ints runs read-only on the realm's own slice.

import "sort"

var gInts = []int{3, 1, 2}

func main(cur realm) {
	f := sort.Ints
	f = sort.Ints
	f(gInts)
	println(gInts)
}

// Error:
// gno.land/r/zrealm_entry_pfunc_sortown's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): x<VPBlock(1,0)>[i<VPBlock(1,1)>]
