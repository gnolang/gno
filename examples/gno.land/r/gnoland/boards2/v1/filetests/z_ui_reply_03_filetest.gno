// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_reply_03_filetest

// Render the "reply to a comment" form: the parent-comment preview is now
// shown inside a labeled <gno-foreign> sandbox (was md.Blockquote, which
// gave no sandboxing).
package z_ui_reply_03_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "test-board"
)

var (
	threadID boards.ID
	replyID  boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boardID := boards2.CreateBoard(cross(cur), boardName, false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Foo", "Body")
	replyID = boards2.CreateReply(cross(cur), boardID, threadID, 0, "Parent comment body")
}

func main() {
	// The ".../{reply}/reply" route renders the parent-preview branch.
	path := boardName + "/" + threadID.String() + "/" + replyID.String() + "/reply"
	content := boards2.Render(path)
	println(strings.Contains(content, "Replying to a comment posted by") &&
		strings.Contains(content, `<gno-foreign label="Quoted comment">`) &&
		strings.Contains(content, "Parent comment body"))
}

// Output:
// true
