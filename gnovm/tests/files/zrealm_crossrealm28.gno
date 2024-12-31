// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

// XXX, consider this
var fs []crossrealm.Fooer

func init() {
	//fs = append(fs, foo{name: "1"}) // type of elem is from external realm
	//s1 := []crossrealm.Fooer{foo{name: "s1"}}
	//r := append(fs, s1)

	r := append(fs, foo{name: "1"}) // type of elem is from external realm, but no panic since no attach
	fs = r                          // panic while attach
}

func main() {
}

// Error:
// cannot attach a reference to an unreal object from an external realm
