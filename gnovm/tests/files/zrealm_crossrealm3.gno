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
// should not happen, mismatched method/receiver realm, &(somevalue<~VPBlock(3,0)>).Modify()/(gno.land/r/demo/tests---------RIDDC8307F7BBE97FDB07B2D0FA6907B8D67517C02A)
