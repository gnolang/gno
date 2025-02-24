// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

// associate to a containing object
type Container struct {
	name string
	b    *crossrealm.Bar
}

var c = Container{name: "local_container"}

func init() {
	b0 := &crossrealm.Bar{A: 1} // this is valid association
	c.b = b0                    // c is real, so panic while association, for b0 is unreal
}

func main() {
	print(".")
}

// Error:
// cannot attach a reference to an unattached object from an external realm
