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
// cannot directly modify readonly tainted object (use a method or crossing function): somevalue<~VPBlock(3,0)>.Field
