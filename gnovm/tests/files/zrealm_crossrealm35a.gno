// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (*foo) String() string { return "hey" }

var f = &foo{"foo"}

var arr interface{}

func main() {
	arr = crossrealm.GetSlice9(f) // return value is unreal, elem is from local realm
	println(".")
}

// Output:
// .
