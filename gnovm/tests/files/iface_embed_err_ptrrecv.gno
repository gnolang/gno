package main

import "filetests/extern/ifaceext"

// T promotes ifaceext.sec through a pointer receiver, so a T value does not
// implement ifaceext.Sec. The diagnostic names the stamped method.
type T struct{ ifaceext.SecPtrImpl }

func main() {
	var _ interface{ ifaceext.Sec } = T{}
	println("unreached")
}

// Error:
// main/iface_embed_err_ptrrecv.gno:10:6-39: main.T does not implement interface {filetests/extern/ifaceext.sec func() int} (method filetests/extern/ifaceext.sec has pointer receiver)

// TypeCheckError:
// main/iface_embed_err_ptrrecv.gno:10:36: cannot use T{} (value of struct type T) as interface{ifaceext.Sec} value in variable declaration: T does not implement interface{ifaceext.Sec} (method sec has pointer receiver)
