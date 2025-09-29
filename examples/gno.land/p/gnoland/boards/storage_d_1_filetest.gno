// PKGPATH: gno.land/r/demo/test
//
// Test removing a board that doesn't exist.
// Board is created but is not saved so it's not stored.
package test

import (
	"gno.land/p/gnoland/boards"
)

func main() {
	board, removed := boards.NewStorage().Remove(404)

	println(removed)
	println(board == nil)
}

// Output:
// false
// true
