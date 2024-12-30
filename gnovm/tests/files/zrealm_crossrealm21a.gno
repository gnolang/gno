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
	c = Container{name: "container", b: b0} // XXX, should panic
}

func main() {
	print(".")
}

// Error:
// cannot attach a value of a type defined by another realm
