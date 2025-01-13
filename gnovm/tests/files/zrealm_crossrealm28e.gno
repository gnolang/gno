// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type foo struct {
	name string
}

func (*foo) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

// XXX, no attach, default value is nil for slice
var fs []crossrealm.Fooer

func init() {
	r := append(fs, &foo{name: "1"})
	// while attaching underlying arrayValue(the object),
	// it panics since the array type if from external realm
	fs = r
}

func main() {
}

// Error:
// cannot attach a reference to an unreal object from an external realm
