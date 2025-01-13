// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

// TODO: consider this and xxx22.gno, xxx21.gno, xxx20.gno
// how about func()int? and func() crossrealm.Fooer?
// check type for param type?
// see also 37a.gno

type Foo struct{}

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(func() crossrealm.Fooer {
		//f1 := Foo{}
		return f
	})
	println("ok")
}

// Output:
// ok
