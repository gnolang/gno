// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

// TODO: think about these two
// array[(struct{("1" string)} gno.land/r/crossrealm_test.foo),(struct{("2" string)} gno.land/r/crossrealm_test.foo),(struct{("3" string)} gno.land/r/crossrealm_test.foo)]
var arr = []crossrealm.Fooer{foo{name: "1"}, foo{name: "2"}, foo{name: "3"}}

//var arr2 = []foo{foo{name: "1"}, foo{name: "2"}, foo{name: "3"}}

func main() {
	println("ok")
}

// Output:
// ok
