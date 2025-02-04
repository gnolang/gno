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

func main() {
	var arr []crossrealm.Fooer // XXX, the array object should be in current realm
	arr = append(arr, foo{name: "1"})
	arr = append(arr, foo{name: "2"})
	arr = append(arr, foo{name: "3"})

	fs := arr[1:]
	crossrealm.SetSlice(fs)
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm
