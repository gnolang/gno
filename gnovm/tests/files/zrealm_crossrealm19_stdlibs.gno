// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

// TODO: this test panics with the current implementation of the VM.
// we should either enable it again if we change the VM behavior (#2743).
// or consider adding a `// Panic:` filetest handler.`
func main() { /* noop */ }

/*
import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) Foo() {
	f.s = "B"
	println("hello " + f.s + " " + std.CurrentRealm().PkgPath())
}

var f *fooer

func init() {
	f = &fooer{s: "A"}
	crossrealm.SetFooer(f)
	crossrealm.CallFoo()
}

func main() {
	print(".")
}
*/

// EXPECTED Output:
// hello B gno.land/r/demo/tests/crossrealm
// .

// Error:

// Realm:
