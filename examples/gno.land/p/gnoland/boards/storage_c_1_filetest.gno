// PKGPATH: gno.land/r/demo/test
//
// Test getting a board that doesn't exist by name.
package test

import (
	"gno.land/p/gnoland/boards"
)

func main() {
	_, found := boards.NewStorage().GetByName("foo")

	println(found)
}

// Output:
// false
