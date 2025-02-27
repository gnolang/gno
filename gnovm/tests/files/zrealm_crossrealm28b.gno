// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var arr [2]crossrealm.Fooer // the element attaching is nil, so it's ok

func main() {
	println("ok")
}

// Output:
// ok
