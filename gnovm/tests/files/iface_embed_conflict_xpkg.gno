package main

import "filetests/extern/ifaceext"

// Both embeds contribute an ifaceext-stamped sec with conflicting signatures.
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
