// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

var root interface{}

// TODO: consider this
// should it be valid to attach
// an object by value with type
// from p? maybe yes, but it should
// be invalid if the type is from
// an another realm.
func main() {
	root = tests.F // already attached in p
	println("done:", root)
}

// Output:
// done: (struct{("p" string)} gno.land/p/demo/tests.Foo)
