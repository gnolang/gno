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
// internal/bytealg/compare_generic.gno:30:5-11: allocation limit exceeded (no GC)
