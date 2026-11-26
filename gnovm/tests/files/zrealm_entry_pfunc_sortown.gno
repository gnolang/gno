// PKGPATH: gno.land/r/zrealm_entry_pfunc_sortown
package zrealm_entry_pfunc_sortown

// A realm calling a stdlib function through a local variable it made in its
// gate holds a local copy, not a call by name: sort.Ints runs read-only.

import "gno.land/r/tests/vm/pfuncvictim"

func main(cur realm) {
	pfuncvictim.SortOwn(cross(cur))
	println(pfuncvictim.Ints())
}

// Error:
// gno.land/r/tests/vm/pfuncvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): x<VPBlock(1,0)>[i<VPBlock(1,1)>]
