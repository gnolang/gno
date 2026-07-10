package main

import "filetests/extern/ifaceext"

// E has no sec at all, so ifaceext.Sec reports its stamped method missing.
// This message was already qualified before diagName existed; the test pins it
// so the shared diagName path cannot regress it.
type E struct{}

func main() {
	var _ interface{ ifaceext.Sec } = E{}
	println("unreached")
}

// Error:
// main/iface_embed_err_missing.gno:11:6-39: main.E does not implement interface {filetests/extern/ifaceext.sec func() int} (missing method filetests/extern/ifaceext.sec)

// TypeCheckError:
// main/iface_embed_err_missing.gno:11:36: cannot use E{} (value of struct type E) as interface{ifaceext.Sec} value in variable declaration: E does not implement interface{ifaceext.Sec} (missing method sec)
