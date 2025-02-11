// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
	b    *bar
}
type bar struct{ name string }

func (fooer) Foo() { println("hello") }

var f = fooer{name: "local_fooer"} // already real
var b = &bar{name: "local_bar"}

func main() {
	f.b = b                      // this makes f dirty, it's real
	crossrealm.SetContainer2(&f) // object DidUpdate is package block
	crossrealm.RunF()
	print(".")
}

// Output:
// hello
// .
