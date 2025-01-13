// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

// XXX, no attach, default value is nil for slice
var local_S []int

func init() {
	local_S = crossrealm.GetSlice10()
}

func main() {
	println("ok")
}

// Output:
// ok
