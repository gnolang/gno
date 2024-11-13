// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var arr [2]crossrealm.Fooer // panic

func main() {
	arr[0] = foo{name: "1"}
	arr[1] = foo{name: "2"}

	fs := arr[1:]
	println("ok")
	crossrealm.SetSlice(fs)
}

// Error:
// should not happen while attempting to attach new real object from external realm
