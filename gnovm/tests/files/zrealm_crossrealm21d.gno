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

func main() {
	f.bar = bar{name: "local_bar"} // this makes f dirty, it's real
	crossrealm.SetContainer2(&f)
	crossrealm.RunF()
	print(".")
}

// Error:
// cannot attach a reference to dirty object from an external realm
