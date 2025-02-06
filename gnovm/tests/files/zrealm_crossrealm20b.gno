// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

// associate to a containing object
type Container struct {
	name string
	b    crossrealm.Bar // panic while attaching(finalizing)
}

var c Container

func init() {
	print(".")
}

func main() {
}

// Error:
// cannot attach a value of a type defined by another realm: struct{(0 int)}
