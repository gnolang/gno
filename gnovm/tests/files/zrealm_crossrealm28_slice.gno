// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var fs []crossrealm.Fooer

// the attachment works cuz
// the object being attached
// is foo{name: "1"}
func init() {
	r := append(fs, foo{name: "1"})
	fs = r // the array is from external realm and unreal
}

func main() {
	println("ok")
}

// Error:
// cannot attach a reference to an unreal object from an external realm
