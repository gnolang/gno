// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello") }

func main() {
	// object attaching to external is main block,
	// f is an unreal value defined w/i this block,
	// and invalid to attach to external, panic.
	f := fooer{name: "local_fooer"}
	crossrealm.SetContainer2(&f)
	print(".")
}

// Error:
// cannot attach a value of a type defined by another realm
