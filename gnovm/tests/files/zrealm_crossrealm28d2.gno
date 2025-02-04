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

var f1 = &foo{name: "1"}
var f2 = &foo{name: "2"}
var f3 = &foo{name: "3"}

func main() {
	var Local_S []crossrealm.Fooer // XXX, the array object should be in current realm
	Local_S = append(Local_S, f1)
	Local_S = append(Local_S, f2)
	Local_S = append(Local_S, f3)

	fs := Local_S[1:]
	crossrealm.SetSlice(fs)
	println("ok")
}

// Output:
// &(ref(f5a516808f8976c33939133293d598ce3bca4e8d:6) gno.land/r/crossrealm_test.foo)
// &(ref(f5a516808f8976c33939133293d598ce3bca4e8d:8) gno.land/r/crossrealm_test.foo)
// ok
