// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/r/demo/tests/crossrealm"

type Foo struct{ num int }

var a interface{}

func main() {
	a = Foo{num: 1}
	f := func() {
		println("---callback to local realm")
	}

	crossrealm.SetClosure(cross, f)

	crossrealm.ExecuteClosure(cross)
}

// Output:
// ---callback to local realm
