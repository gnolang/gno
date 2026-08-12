// Multi-level cross-package embedding: ifacemid.Mid embeds ifaceext.Sec (an
// unexported sec() owned by ifaceext). Embedding Mid here must keep sec's
// origin package as ifaceext through the second hoist (ifaceext -> ifacemid ->
// main), so interface{ ifacemid.Mid } and interface{ ifaceext.Sec } have the
// same method set {ifaceext.sec} and one TypeID. Guards the origin-preserving
// branch in flattenInterfaceMethods (m.PkgPath kept over the embed's PkgPath).
package main

import (
	"filetests/extern/ifaceext"
	"filetests/extern/ifacemid"
)

func main() {
	var a, b interface{} = struct{ X interface{ ifacemid.Mid } }{}, struct{ X interface{ ifaceext.Sec } }{}
	println("midSec==sec:", a == b)
}

// Output:
// midSec==sec: true
