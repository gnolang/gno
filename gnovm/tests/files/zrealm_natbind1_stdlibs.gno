// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

func main() {
	crossing()

	println(std.ChainDomain())
}

// Output:
// gno.land

// Realm:
// switchrealm["gno.land/r/test"]
