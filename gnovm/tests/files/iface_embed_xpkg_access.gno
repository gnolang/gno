// Selecting a cross-package unexported method is rejected at compile time
// even after flattening hoists it into a locally-spelled interface: the
// access gate keys on the method's origin package (ifaceext), not the
// enclosing interface's (main). Matches Go and pre-flattening master.
package main

import "filetests/extern/ifaceext"

func main() {
	var x interface{ ifaceext.Sec } = ifaceext.SecImpl{}
	println(x.sec())
}

// Error:
// main/iface_embed_xpkg_access.gno:11:10-15: cannot access interface {filetests/extern/ifaceext.sec func() int}.sec from main

// TypeCheckError:
// main/iface_embed_xpkg_access.gno:11:12: x.sec undefined (cannot refer to unexported method sec)
