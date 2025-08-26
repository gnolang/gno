// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/demo/tests"
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
// invalid call: &(somevalue<~VPBlock(3,0)>).Modify(), method realm: RID0FFE7732B4D549B4CF9EC18BD68641CD2C75AD0A does not match receiver realm: RIDDC8307F7BBE97FDB07B2D0FA6907B8D67517C02A
