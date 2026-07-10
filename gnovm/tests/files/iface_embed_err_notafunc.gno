package main

import "filetests/extern/ifaceext"

// N promotes ifaceext's sec field, so the lookup finds a non-function where
// ifaceext.Sec wants a method.
type N struct{ ifaceext.FieldSec }

func main() {
	var _ interface{ ifaceext.Sec } = N{}
	println("unreached")
}

// Error:
// main/iface_embed_err_notafunc.gno:10:6-39: main.N does not implement interface {filetests/extern/ifaceext.sec func() int} (wrong type for method filetests/extern/ifaceext.sec)

// TypeCheckError:
// main/iface_embed_err_notafunc.gno:10:36: cannot use N{} (value of struct type N) as interface{ifaceext.Sec} value in variable declaration: N does not implement interface{ifaceext.Sec} (N.sec is a field, not a method)
