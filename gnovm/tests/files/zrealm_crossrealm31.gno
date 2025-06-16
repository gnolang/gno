// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	rb "gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	fooer := rb.New(cross)
	fooer.SetS("B")
	rb.Foo2(cross)
}

// Output:
// hello A
