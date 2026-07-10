package main

import "filetests/extern/ifaceext"

// The panic must name the stamped sec, not a bare one.
type I interface {
	ifaceext.Sec
	ifaceext.SecStr
}

func main() {
	var _ I
	println("unreached")
}

// Error:
// main/iface_embed_conflict_xpkg.gno:6:8-9:2: duplicate method filetests/extern/ifaceext.sec with conflicting types in interface

// TypeCheckError:
// main/iface_embed_conflict_xpkg.gno:8:2: duplicate method sec; main/iface_embed_conflict_xpkg.gno:7:2: 	other declaration of method sec
