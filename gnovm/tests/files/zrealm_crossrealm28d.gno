// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (*foo) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

func main() {
	var Local_S []crossrealm.Fooer // XXX, the array object should be in current realm
	Local_S = append(Local_S, &foo{name: "1"})
	Local_S = append(Local_S, &foo{name: "2"})
	Local_S = append(Local_S, &foo{name: "3"})

	fs := Local_S[1:]
	crossrealm.SetSlice(fs)
	println("ok")
}

// Output:
// &(struct{("2" string)} gno.land/r/crossrealm_test.foo)
// &(struct{("3" string)} gno.land/r/crossrealm_test.foo)
// ok
