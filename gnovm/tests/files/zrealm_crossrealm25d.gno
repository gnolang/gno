// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

// pass pointer to an embedded object to external realm?

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

type localContainer struct {
	name string
	ff   fooer
}

var lc localContainer

func init() {
	f = fooer{name: "local_fooer_global"}
	lc = localContainer{name: "local_container", ff: f} // attach
}

func main() {
	crossrealm.AttachContainer()   // this attaches container first
	crossrealm.SetContainer(lc.ff) // this should not panic, lc.ff is not same with f(gloabl)
	print(".")
}
