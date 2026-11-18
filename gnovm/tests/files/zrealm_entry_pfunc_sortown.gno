// PKGPATH: gno.land/r/zrealm_entry_pfunc_sortown
package zrealm_entry_pfunc_sortown

// A realm calling a stdlib function through a local variable it made in its
// gate acts like calling it by name: sort.Ints writes the realm's own data.

import "gno.land/r/tests/vm/pfuncvictim"

func main(cur realm) {
	pfuncvictim.SortOwn(cross(cur))
	println(pfuncvictim.Ints())
}

// Output:
// slice[(1 int),(2 int),(3 int)]
