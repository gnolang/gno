// Test creating a thread without a parent board.
package main

import (
	"gno.land/p/gnoland/boards"
)

func main() {
	boards.MustNewThread(nil, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
}

// Error:
// thread requires a parent board
