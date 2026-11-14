// PKGPATH: gno.land/r/zrealm_entry_pfunc_sortown
package zrealm_entry_pfunc_sortown

// A realm calling a stdlib function through a local variable on its own data
// runs it view-only too; calling sort.Ints by name writes.

import "gno.land/r/tests/vm/pfuncvictim"

func main(cur realm) {
	pfuncvictim.SortOwn()
}

// Error:
// illegal conversion of readonly or externally stored value
