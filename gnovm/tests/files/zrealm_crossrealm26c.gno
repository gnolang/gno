// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{ name string }

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

func main() {
	crossrealm.AttachContainer()             // this attaches container first
	var f fooer = fooer{name: "local_fooer"} // this is not attached, should panic, even not escape
	crossrealm.SetContainer(f)
	print(".")
}
