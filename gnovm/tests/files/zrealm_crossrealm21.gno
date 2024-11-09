// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 crossrealm.Bar // run declarations

func init() {
	b0 = crossrealm.Bar{A: 1}
}

func main() {
	print(".")
}

// XXX, this should fail while trying to attach
// value of type defined in another realm.
// NOTE: the check can be in runtime or in preprocess(by simply checkin edge cases).

// Error:
// !!!cross realm while attach object
