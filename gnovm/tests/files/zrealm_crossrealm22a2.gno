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

type local_S struct{ name string }

var root interface{}

func main() { // the block is attached to current realm
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(func() crossrealm.Fooer { // this funcLit is attached to external realm
		return f
	})
	crossrealm.CallFooerGetterFoo()
	println("ok")

	root = local_S{name: "root"}
	println("ok2")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// ok
// ok2
