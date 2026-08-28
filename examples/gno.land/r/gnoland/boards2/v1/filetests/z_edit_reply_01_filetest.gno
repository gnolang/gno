// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_edit_reply_01_filetest

package z_edit_reply_01_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	body          = "Test reply"
	path          = "test-board/1/2"
)

var (
	bid      boards.ID
	tid, rid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")

	// Create a reply and a sub reply
	parentRID := boards2.CreateReply(cross(cur), bid, tid, 0, "Parent")
	rid = boards2.CreateReply(cross(cur), bid, tid, parentRID, "Child")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditReply(cross(cur), bid, tid, rid, body)

	// Render content must contain the modified reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> > "+body))
}

// Output:
// true
