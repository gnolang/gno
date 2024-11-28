// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm"

type LocalBar crossrealm.Bar

var lb LocalBar

func main() {
	println(lb)
}
