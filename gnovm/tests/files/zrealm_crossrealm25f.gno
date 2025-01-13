// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct{ name string }

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer
var root interface{}

func init() {
	f = fooer{name: "local fooer"}
	g := f
	root = g
}

func main() {
	print(".")
}

// Output:
// .
