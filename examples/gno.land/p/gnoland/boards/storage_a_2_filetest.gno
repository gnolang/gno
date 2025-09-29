// PKGPATH: gno.land/r/demo/test
//
// Test saving using a nil board in the storage.
package test

import (
	"gno.land/p/gnoland/boards"
)

var storage = boards.NewStorage()

func main() {
	err := storage.Save(nil)
	if err == nil {
		panic("espected storage save to return an error")
	}

	println(err.Error())
}

// Output:
// save failed, board is nil
