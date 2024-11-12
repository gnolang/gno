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
	f    fooer
}

var lc localContainer

func init() {
	lc = localContainer{name: "local_container", f: fooer{name: "local_fooer"}} // attach
}

func main() {
	crossrealm.SetContainer(lc.f)
	print(".")
}

// Error:
// should not happen while attempting to attach objects by value from external realm
