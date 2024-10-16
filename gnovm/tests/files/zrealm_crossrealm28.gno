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

var fs []crossrealm.Fooer

func init() {
	fs = append(fs, foo{name: "1"})
	fs = append(fs, foo{name: "2"})
	fs = append(fs, foo{name: "3"})
}

func main() {
	println("ok")
	crossrealm.SetSlice(fs)
}
