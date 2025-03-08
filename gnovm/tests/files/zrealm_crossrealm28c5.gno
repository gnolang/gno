// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

func main() {
	var f1 = foo{name: "1"}
	var f2 = foo{name: "2"}
	var f3 = foo{name: "3"}

	var arr [3]crossrealm.Fooer // unreal array

	arr[0] = &f1
	arr[1] = &f2
	arr[2] = &f3

	crossrealm.SetArray(arr)
	println("ok")
}

// Error:
// cannot attach a reference to an unattached object from an external realm
