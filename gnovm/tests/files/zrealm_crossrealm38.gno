// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm"

var f func() string

func main() {
	f = crossrealm.GetFunc()
	println(f())
}

// Output:
// a
