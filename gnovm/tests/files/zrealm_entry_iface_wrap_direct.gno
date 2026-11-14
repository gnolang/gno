// PKGPATH: gno.land/r/zrealm_entry_iface_wrap_direct
// The realm writes its own object through a library it imports.
package zrealm_entry_iface_wrap_direct

import "gno.land/r/tests/vm/ifacewrapuse"

func main(cur realm) {
	println(ifacewrapuse.Direct())
}

// Output:
// direct
