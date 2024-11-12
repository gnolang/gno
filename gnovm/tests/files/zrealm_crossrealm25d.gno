// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

type localContainer struct {
	name string
	ff   fooer
}

var lc localContainer

func init() {
	f := fooer{name: "local_fooer_global"}
	lc = localContainer{name: "local_container", ff: f} // attach
}

func main() {
	crossrealm.SetContainer(lc.ff)
	print(".")
}

// Error:
// should not happen while attempting to attach objects by value from external realm
