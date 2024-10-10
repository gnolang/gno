// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

// associate to a containing object
type Container struct {
	name string
	b    crossrealm.Bar
}

var c Container

func init() {
	b0 := crossrealm.Bar{A: 1}              // this is valid association
	c = Container{name: "container", b: b0} // XXX, check recursively while finalize here
}

func main() {
	print(".")
}

// XXX, this should fail while trying to attach
// value of type defined in another realm.
// NOTE: the check can be in runtime or in preprocess(by simply checkin edge cases).

// Error:
