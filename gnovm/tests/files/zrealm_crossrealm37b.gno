// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var local_m crossrealm.MyMap // default value is nil

func main() {
	local_m = crossrealm.GetMap3()
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm
