// PKGPATH: gno.land/r/demo/test
//
// Test creating a thread reply using an invalid creator address.
package test

import (
	"gno.land/p/gnoland/boards"
)

var parent boards.Post

func main() {
	boards.MustNewReply(&parent, "foo", "Message")
}

// Error:
// invalid reply creator address: foo
