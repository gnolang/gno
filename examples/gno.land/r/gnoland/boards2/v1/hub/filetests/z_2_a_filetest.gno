// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_2_a_filetest

// Test getting boards when no boards exist
package z_2_a_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	boards := hub.GetBoards(0, boards2.BoardCount())
	println(len(boards))
}

// Output:
// 0
