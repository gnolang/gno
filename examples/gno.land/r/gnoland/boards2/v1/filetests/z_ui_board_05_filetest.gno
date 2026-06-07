// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_05_filetest

// An invalid ?page= (out of range, zero, negative, or non-numeric) clamps to
// the last valid page instead of aborting the render or echoing a broken
// "page N of M" picker (see newClampedPager). Notably a NEGATIVE page is not
// rejected by pager.New, so it must be caught by the clamp.
package z_ui_board_05_filetest

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "test-board"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boardID := boards2.CreateBoard(cross(cur), boardName, false, false)
	// pageSizeDefault is 6, so 7 threads span 2 pages; the last page is 2.
	for i := 0; i < 7; i++ {
		boards2.CreateThread(cross(cur), boardID, "T", "Body")
	}
}

func main() {
	// Each value is invalid: too high, zero, negative, non-numeric. All must
	// clamp to the last page ("page 2 of 2"); if any slipped through, the
	// picker would echo the bad number (e.g. "page -5 of 2") and this fails.
	ok := true
	for _, q := range []string{"?page=99", "?page=0", "?page=-5", "?page=abc"} {
		if !strings.Contains(boards2.Render(boardName+q), "page 2 of 2") {
			ok = false
		}
	}
	println(ok)
}

// Output:
// true
