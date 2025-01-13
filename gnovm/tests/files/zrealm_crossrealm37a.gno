// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var local_m map[string]int

var f = func(m map[string]int) {
	local_m = m
}

func main() {
	crossrealm.GetMap2(f)
	println("ok")
}

// Output:
// ok
