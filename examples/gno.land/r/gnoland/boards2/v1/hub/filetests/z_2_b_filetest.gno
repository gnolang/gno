// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_2_b_filetest

// Test getting boards
package z_2_b_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boards2.CreateBoard(cross(cur), "aaa123", false, false)
	boards2.CreateBoard(cross(cur), "bbb123", false, false)
	boards2.CreateBoard(cross(cur), "ccc123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	boards := hub.GetBoards(0, boards2.BoardCount())

	for _, b := range boards {
		println(b.ID, b.Name)
	}
}

// Output:
// 1 aaa123
// 2 bbb123
// 3 ccc123
