// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var arr = []crossrealm.Fooer{foo{name: "1"}, foo{name: "2"}, foo{name: "3"}}

func main() {
	println("ok")
}

// Output:
// ok
