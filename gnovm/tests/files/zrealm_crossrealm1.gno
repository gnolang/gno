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
// cannot directly modify readonly tainted object (call a crossing function of the owning realm): somevalue<~VPBlock(3,0)>.Field

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm1.gno:11:5: cannot hold a value of gno.land/r/tests/vm.TestRealmObject outside its realm (use a pointer)
