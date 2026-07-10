package main

import "filetests/extern/ifaceext"

// W promotes ifaceext's sec() string, so it cannot satisfy ifaceext.Sec,
// whose sec returns int. The diagnostic names the stamped method.
type W struct{ ifaceext.SecStrImpl }

func main() {
	var _ interface{ ifaceext.Sec } = W{}
	println("unreached")
}

// Error:
// main/iface_embed_err_wrongtype.gno:10:6-39: main.W does not implement interface {filetests/extern/ifaceext.sec func() int} (wrong type for method filetests/extern/ifaceext.sec)

// TypeCheckError:
// main/iface_embed_err_wrongtype.gno:10:36: cannot use W{} (value of struct type W) as interface{ifaceext.Sec} value in variable declaration: W does not implement interface{ifaceext.Sec} (wrong type for method sec)
// 		have sec() string
// 		want sec() int
