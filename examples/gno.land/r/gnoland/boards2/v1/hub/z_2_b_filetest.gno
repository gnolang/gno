// Test getting boards
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boards2.CreateBoard(cross, "aaa123", false, false)
	boards2.CreateBoard(cross, "bbb123", false, false)
	boards2.CreateBoard(cross, "ccc123", false, false)
}

func main() {
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
