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
	var arr []crossrealm.Fooer // unreal array with type from external realm
	arr = append(arr, foo{name: "1"})
	arr = append(arr, foo{name: "2"})
	arr = append(arr, foo{name: "3"})

	fs := arr[1:]
	// this works cuz it's not crossing realm,
	// even underlying array is unreal
	crossrealm.SetSlice(fs) // valid
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm: struct{("1" string)}
