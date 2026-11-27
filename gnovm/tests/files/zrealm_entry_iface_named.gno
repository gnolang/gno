// PKGPATH: gno.land/r/zrealm_entry_iface_named
// A realm that imports a library trusts it: the library's method writes for the
// realm when named, and when reached through the realm's interface.
package zrealm_entry_iface_named

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifaceimporter"
)

func main(cur realm) {
	ifaceimporter.UseNamed(cross(cur))
	println("named:", ifaceimporter.Read())
	ifaceimporter.Use(cross(cur), ifaceattack.EvilV{})
	println("iface:", ifaceimporter.Read())
}

// Output:
// named: pwnd-via-iface
// iface: pwnd-via-iface
