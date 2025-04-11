// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

func main() {
	println(runtime.ChainDomain())
}

// Output:
// gno.land

// Realm:
// switchrealm["gno.land/r/test"]
