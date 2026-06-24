// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_2_a_filetest

// Test getting boards when no boards exist
package z_hub_2_a_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	boards := boards2.GetBoards(0, boards2.BoardCount())
	println(len(boards))
}

// Output:
// 0
