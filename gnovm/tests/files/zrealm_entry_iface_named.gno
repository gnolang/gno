// PKGPATH: gno.land/r/zrealm_entry_iface_named
// Named in the realm's own source, the library method writes for it.
package zrealm_entry_iface_named

import (
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.UseNamed()
	println("victim:", pfuncvictim.Read())
}

// Output:
// victim: pwnd-via-iface
