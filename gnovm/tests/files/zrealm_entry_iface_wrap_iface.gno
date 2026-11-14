// PKGPATH: gno.land/r/zrealm_entry_iface_wrap_iface
// The realm writes its own object through a library it imports.
package zrealm_entry_iface_wrap_iface

import "gno.land/r/tests/vm/ifacewrapuse"

func main(cur realm) {
	println(ifacewrapuse.Iface())
}

// Output:
// iface
