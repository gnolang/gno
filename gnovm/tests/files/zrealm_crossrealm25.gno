// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

// When an association occurs by reference to an object already
// attached to an external realm, no attachment occurs. For example,
// pointers to external realm attached objects may be associated
// with a realm object (whether attached or not), such as assigned
// to the field of a struct, but this does not attach the external
// realm object to the container, in this case the struct.
func init() {
	f = fooer{}
}

func main() {
	// XXX, this associate a pointer to a value(already attached to local realm)
	// to external realm, no attachment happens to the external realm.
	crossrealm.AttachContainer() // this attach container first
	crossrealm.SetContainer(&f)
	print(".")
}

// Output:
// .

// Error:

// Realm:
