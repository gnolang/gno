// A cross-package unexported method promoted through a struct-embedded
// INTERFACE field is callable through the interface: satisfaction admits the
// value (identity matches via the method's origin package), and runtime
// dispatch must locate the same method without re-applying the compile-time
// access gate (Go dispatches via the method set with no access check).
package main

import "filetests/extern/ifaceext"

type S struct{ ifaceext.Sec }

func main() {
	var v ifaceext.Sec = S{Sec: ifaceext.SecImpl{}} // satisfaction: ok
	println(ifaceext.UseSec(v))                     // dispatch from origin pkg
}

// Output:
// 7
