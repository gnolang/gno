// MAXALLOC: 10000

package main

import (
	"gno.land/r/demo/boards" // not enough memory to import
)

func main() {
	_ = boards.CreateBoard(cross, "testboard")
}

// Error:
// main/alloc_10.gno:6:2-26: allocation limit exceeded
