// PKGPATH: gno.land/r/demo/test
//
// Test getting a board that doesn't exist.
package test

import (
	"gno.land/p/gnoland/boards"
)

func main() {
	_, found := boards.NewStorage().Get(404)

	println(found)
}

// Output:
// false
