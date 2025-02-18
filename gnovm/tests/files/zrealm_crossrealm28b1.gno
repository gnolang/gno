// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var arr [2]crossrealm.Foo2 // object attaching is zero value of struct, panic

func main() {
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm
