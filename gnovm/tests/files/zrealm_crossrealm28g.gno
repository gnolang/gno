// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

// XXX, no attach, default value is nil for slice
var local_S []int

var f = func(s1 []int) {
	local_S = s1
}

func init() {
	crossrealm.GetSlice11(f)
}

func main() {
	println("ok")
}

// Output:
// ok
