// PKGPATH: gno.land/r/test/alloc_10
// MAXALLOC: 10000

package alloc_10

import (
	"gno.land/r/archive/boards"
)

func main(cur realm) {
	// not enough memory to run.
	_ = boards.CreateBoard(cross(cur), "testboard")
}

// Error:
// allocation limit exceeded
