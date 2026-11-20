// PKGPATH: gno.land/r/zrealm_entry_pfunc_named
package zrealm_entry_pfunc_named

// The victim's own source names the /p/ function, so it writes as the victim.

import "gno.land/r/tests/vm/pfuncvictim"

func main(cur realm) {
	pfuncvictim.ApplyNamed(cross(cur))
	println("victim:", pfuncvictim.Read())
}

// Output:
// victim: pwnd-via-apply
