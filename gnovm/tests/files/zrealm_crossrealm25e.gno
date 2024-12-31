// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

type localContainer struct {
	name string
	f    fooer
}

var lc localContainer
var ff fooer

func main() {
	ff = fooer{name: "local_fooer"} // ref 1, unreal
	lc = localContainer{name: "local_container", f: ff}
	crossrealm.SetContainer(lc.f)
	print(".")
}

// Error:
// cannot attach objects by value from external realm
