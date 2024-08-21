// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

var (
	a int    = 10
	f *fooer = &fooer{}
)

func (f *fooer) Foo() { a *= 2 }

func init() {
	a++                    // a=11
	crossrealm.SetFooer(f) // a=11
	a++                    // a=12
}

func main() {
	a++                  // a=13
	crossrealm.CallFoo() // a=26
	a++                  // a=27
	println(a)
}

// Output:
// 27

// Error:

// Realm:
