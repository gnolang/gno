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

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f *fooer

func init() {
	f = &fooer{}
	crossrealm.SetFooer(f)
	crossrealm.CallFoo()
}

func main() {
	print(".")
}
*/

// EXPECTED Output:
// hello gno.land/r/crossrealm_test
// .

// Error:

// Realm:
