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

// Error:
// gno.land/r/crossrealm_test/zrealm_crossrealm31.gno:9:11-17: name New not declared

// TypeCheckError:
// gno.land/r/crossrealm_test/zrealm_crossrealm31.gno:9:14: undefined: rb.New; gno.land/r/crossrealm_test/zrealm_crossrealm31.gno:11:5: undefined: rb.Foo2
