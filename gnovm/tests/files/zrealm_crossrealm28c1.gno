// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

// should panic, underlying array with external type is unreal
var arr = []crossrealm.Fooer{foo{name: "1"}, foo{name: "2"}, foo{name: "3"}}

func main() {
}

// Error:
// cannot attach a reference to an unreal object from an external realm
