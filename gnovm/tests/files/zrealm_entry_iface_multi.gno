// PKGPATH: gno.land/r/zrealm_entry_iface_multi
package zrealm_entry_iface_multi

// The realm's own buffer written through a stdlib writer it built still writes.
import "gno.land/r/tests/vm/ifacevictim"

func main(cur realm) {
	println(ifacevictim.WriteMulti())
	println("victim:", ifacevictim.Read())
}

// Output:
// hi
// victim: victim-orig
