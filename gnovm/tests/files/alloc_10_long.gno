// MAXALLOC: 10000

package main

import (
	"gno.land/r/archive/boards" // not enough memory to import
)

func main() {
	_ = boards.CreateBoard(cross, "testboard")
}

// Error:
// main/alloc_10_long.gno:6:2-29: should not happen, allocation limit exceeded while gc.
