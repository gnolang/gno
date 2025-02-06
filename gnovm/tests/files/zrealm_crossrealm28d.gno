// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

func main() {
	var Local_S []crossrealm.Fooer
	Local_S = append(Local_S, &foo{name: "1"})
	Local_S = append(Local_S, &foo{name: "2"})
	Local_S = append(Local_S, &foo{name: "3"})

	fs := Local_S[1:]
	crossrealm.SetSlice(fs)
	println("ok")
}

// Error:
// cannot attach a reference to an unreal object from an external realm: heapitem((struct{("2" string)} gno.land/r/crossrealm_test.foo))
