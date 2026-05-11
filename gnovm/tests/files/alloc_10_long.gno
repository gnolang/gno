// MAXALLOC: 10000

package main

import (
	"gno.land/r/archive/boards"
)

func main() {
	// not enough memory to run.
	_ = boards.CreateBoard(cross, "testboard")
}

// Error:
// allocation limit exceeded
