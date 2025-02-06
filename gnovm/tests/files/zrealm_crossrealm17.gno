// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type container struct{ f *fooer }

func (container) Foo() { println("hello") }

type fooer struct{ name string }

var c = &container{}

func main() {
	// the object attaching is heapItemValue, not escaped
	// while attempting to attach it to external realm,
	// check recursively.
	// also see xxx_crossrealm21.gno
	ff := &fooer{name: "local_foo"}
	c.f = ff
	crossrealm.SetFooer(c)
	crossrealm.CallFooerFoo()
	print(".")
}

// Error:
// cannot attach a reference to dirty object from an external realm
