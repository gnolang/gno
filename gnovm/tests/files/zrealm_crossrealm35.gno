// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (*foo) String() string { return std.CurrentRealm().PkgPath() }

var f = &foo{"foo"}

var arr interface{}

func main() {
	arr = crossrealm.GetSlice8(f)
	println(".")
}

// Error:
// cannot attach a value of a type defined by another realm
