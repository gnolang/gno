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
	arr = crossrealm.GetSlice8(f) // real array value from external realm
	println(".")
}

// Error:
// cannot attach a value of a type defined by another realm: array[(&(ref(f5a516808f8976c33939133293d598ce3bca4e8d:4) gno.land/r/crossrealm_test.foo) *gno.land/r/crossrealm_test.foo)]
