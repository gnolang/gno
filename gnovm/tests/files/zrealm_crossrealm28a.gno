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

var arr []crossrealm.Fooer

func main() {
	arr = append(arr, foo{name: "1"})
	arr = append(arr, foo{name: "2"})
	arr = append(arr, foo{name: "3"})

	println(".")
}

// Output:
// .
