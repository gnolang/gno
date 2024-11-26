// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

// XXX, consider this
var fs []crossrealm.Fooer

func init() {
	fs = append(fs, foo{name: "1"})
	fs = append(fs, foo{name: "2"})
}

func main() {
	crossrealm.SetSlice(fs)
	println("ok")
}

// Output:
// (struct{("1" string)} gno.land/r/crossrealm_test.foo)
// (struct{("2" string)} gno.land/r/crossrealm_test.foo)
// ok
