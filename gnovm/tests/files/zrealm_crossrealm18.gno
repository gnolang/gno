// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello") }

var f crossrealm.Fooer = crossrealm.SetFooer(&fooer{})

func init() {
	crossrealm.CallFoo()
}

func main() {
	crossrealm.CallFoo()
	print(".")
}

// Output:
// hello
// hello
// .

// Error:

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
