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

var c = &Container{name: "local_container", b: &crossrealm.Bar{A: 1}}

func main() {
	print(".")
}

// Error:
// cannot attach a reference to an unreal object from an external realm: heapitem((struct{(1 int)} gno.land/r/demo/tests/crossrealm/struct.Bar))
