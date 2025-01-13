// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

func main() {
	println(std.GetChainDomain())
}

// Output:
// tests.gno.land

// Realm:
// switchrealm["gno.land/r/test"]
