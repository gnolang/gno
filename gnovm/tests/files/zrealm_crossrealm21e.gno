// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
	bar
}
type bar struct{ name string }

func (fooer) Foo() { println("hello") }

var f = fooer{name: "local_fooer"}
var b = bar{name: "local_bar"}

func main() {
	f.bar = b                    // this makes f dirty
	crossrealm.SetContainer2(&f) // can not attach, f may change, so currently it's not determined
	crossrealm.RunF()
	print(".")
}

// Output:
// hello
// .
