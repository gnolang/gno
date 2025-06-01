// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

func main(cur realm) {
	println(std.ChainDomain())
}

// Output:
// gno.land

// Realm:
// finalizerealm["gno.land/r/test"]
