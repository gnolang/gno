// Test getting flags from a comment
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	moderator         = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
)

var comment hub.Comment

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boardID := boards2.CreateBoard(cross, "test123", false, false)
	threadID := boards2.CreateThread(cross, boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross, boardID, threadID, 0, "Comment")

	// Invite member
	boards2.InviteMember(cross, boardID, moderator, "moderator")

	// Update flagging threshold to two flags
	boards2.SetFlaggingThreshold(cross, boardID, 2)

	// Add first flag
	boards2.FlagReply(cross, boardID, threadID, commentID, "Reason 1")

	// Add second flag
	testing.SetRealm(testing.NewUserRealm(moderator))
	boards2.FlagReply(cross, boardID, threadID, commentID, "Reason 2")

	// Get readonly comment
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comment, _ = hub.GetComment(uint64(boardID), uint64(threadID), uint64(commentID))
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	flags := hub.GetFlags(comment.BoardID, comment.ThreadID, comment.ID, 0, comment.FlagCount)

	for _, f := range flags {
		println(f.User, f.Reason)
	}
}

// Output:
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq Reason 1
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj Reason 2
