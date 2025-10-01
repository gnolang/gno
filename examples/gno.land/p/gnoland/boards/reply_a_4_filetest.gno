// PKGPATH: gno.land/r/demo/test
//
// Test creating a thread reply using an empty body.
package test

import (
	"gno.land/p/gnoland/boards"
)

var parent boards.Post

func main() {
	boards.MustNewReply(&parent, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "")
}

// Error:
// reply body is required
