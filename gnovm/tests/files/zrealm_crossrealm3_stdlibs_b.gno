// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

// TODO: this test panics with the current implementation of the VM.
// we should either enable it again if we change the VM behavior (#2743).
// or consider adding a `// Panic:` filetest handler.`
func main() { /* noop */ }

/*
import (
	"gno.land/r/demo/tests"
)


// NOTE: it is valid to persist external realm types.
var somevalue tests.TestRealmObject

func init() {
	somevalue.Field = "test"
}

func main() {
	// NOTE: it is valid to modify it using an external realm function,
	// because the receiver makes the function run under this realm.
	somevalue.Modify()
	println(somevalue)
	// we can even modify it directly.
	somevalue.Field = "test_modified_directly"
	println(somevalue)
}
*/

// EXPECTED Output:
// (struct{("test_modified" string)} gno.land/r/demo/tests.TestRealmObject)
// (struct{("test_modified_directly" string)} gno.land/r/demo/tests.TestRealmObject)
