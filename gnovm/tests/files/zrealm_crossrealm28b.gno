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

var arr [2]crossrealm.Fooer // panic while attach zero value

func main() {
}

// Error:
// cannot attach a value of a type defined by another realm
