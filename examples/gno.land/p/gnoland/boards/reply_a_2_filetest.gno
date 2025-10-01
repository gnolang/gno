// PKGPATH: gno.land/r/demo/test
//
// Test creating a thread reply using a nil parent.
package test

import (
	"gno.land/p/gnoland/boards"
)

func main() {
	boards.MustNewReply(nil, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Message")
}

// Error:
// reply requires a parent thread or reply
