// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello") }

var f = fooer{name: "local_fooer"}

func main() {
	crossrealm.SetContainer2(&f) // the object attached is package block, escaped, so this works.
	crossrealm.RunF()
	print(".")
}

// Output:
// hello
// .
