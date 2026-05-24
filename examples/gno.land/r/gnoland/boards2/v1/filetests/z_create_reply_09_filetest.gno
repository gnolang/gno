// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_09_filetest

package z_create_reply_09_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	path            = "test-board/1/2"
	comment         = "Second comment"
)

var (
	bid      boards.ID
	tid, rid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "First comment")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	rid2 := boards2.CreateReply(cross(cur), bid, tid, rid, comment)

	// Ensure that returned ID is right
	println(rid2 == 3)

	// Render content must contain the sub-reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> > "+comment+"\n"))
}

// Output:
// true
// true
