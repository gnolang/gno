// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

func main() {
	// object attaching to external is main block,
	// f is a value defined w/i this block,
	// and invalid to attach to external, panic.
	f := fooer{name: "local_fooer"}
	crossrealm.SetContainer2(&f)
	print(".")
}

// Error:
// cannot attach a reference to an unreal object from an external realm: struct{("local_fooer" string)}
