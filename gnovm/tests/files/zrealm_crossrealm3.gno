// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// NOTE: it is valid to persist external realm types.
var somevalue tests.TestRealmObject

func init() {
	somevalue.Field = "test"
}

func main(cur realm) {
	// NOTE: it is also valid to modify it using an external realm function.
	somevalue.Modify()
	println(somevalue)
}

// Error:
// invalid call: &(somevalue<~VPBlock(3,0)>).Modify(), method realm: RID87878546FB8F1E4C09BE337A32C14E2E64A869D3 does not match receiver realm: RIDDC8307F7BBE97FDB07B2D0FA6907B8D67517C02A
