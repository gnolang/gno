// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type container struct{ f *fooer }

func (container) Foo() { println("hello container " + std.CurrentRealm().PkgPath()) }

type fooer struct{ name string }

var f *fooer
var c = &container{}

func main() {
	// the object attaching is heapItemValue, not escaped
	// while attempting to attach it to external realm,
	// check recursively.
	// also see xxx_crossrealm21.gno
	f = &fooer{name: "local_foo"}
	c.f = f
	crossrealm.SetFooer(c)
	crossrealm.CallFooerFoo()
	print(".")
}

// Error:
// cannot attach a reference to an unreal object from an external realm
