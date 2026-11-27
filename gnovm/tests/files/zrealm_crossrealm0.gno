// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// Regression: a foreign-realm-typed package var carries the type's
// declared PkgID (type-driven stamping). /r/crossrealm cannot
// directly write fields of an /r/tests/vm-stamped struct — the
// readonly gate fires.
var somevalue tests.TestRealmObject

func main(cur realm) {
	somevalue.Field = "test"
}

// Error:
// cannot allocate gno.land/r/tests/vm.TestRealmObject in realm gno.land/r/crossrealm

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm0.gno:12:5: cannot hold a value of gno.land/r/tests/vm.TestRealmObject outside its realm (use a pointer)
