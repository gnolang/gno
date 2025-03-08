// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var f1 = foo{name: "1"}
var f2 = foo{name: "2"}
var f3 = foo{name: "3"}

func main() {
	var arr []crossrealm.Fooer // unreal array
	arr = append(arr, f1)
	arr = append(arr, f2)
	arr = append(arr, f3)

	crossrealm.SetSlice(arr)
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm
