// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// Same as crossrealm0 but the field write happens during package
// init. The init frame runs under /r/crossrealm's realm, so the
// write to an /r/tests/vm-stamped value is still rejected.
var somevalue tests.TestRealmObject

func init() {
	somevalue.Field = "test"
}

func main(cur realm) {}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): somevalue<~VPBlock(3,0)>.Field
