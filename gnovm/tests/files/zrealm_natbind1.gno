// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

func main(cur realm) {
	println(runtime.ChainDomain())
}

// Output:
// gno.land

// Realm:
// finalizerealm["gno.land/r/test"]
