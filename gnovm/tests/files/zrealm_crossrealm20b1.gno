// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

// associate to a containing object
type Container struct {
	name string
	b    crossrealm.Bar
}

var c *Container

func init() {
	c = &Container{name: "Container", b: crossrealm.Bar{A: 1}}
	println(c.b)
}

func main() {
}

// Error:
// cannot attach a value of a type defined by another realm
