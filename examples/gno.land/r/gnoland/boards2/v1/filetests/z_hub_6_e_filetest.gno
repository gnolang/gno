// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_6_e_filetest

// Test getting flags from a nested reply
package z_hub_6_e_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	admin             = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	moderator         = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
)

var (
	boardID  boards.ID
	threadID boards.ID
	replyID  boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")
	replyID = boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply")

	// Invite members
	boards2.InviteMember(cross(cur), boardID, admin, "admin")
	boards2.InviteMember(cross(cur), boardID, moderator, "moderator")

	// Update flagging threshold to two flags
	boards2.SetFlaggingThreshold(cross(cur), boardID, 2)

	// Flag the nested reply
	testing.SetRealm(testing.NewUserRealm(admin))
	boards2.FlagReply(cross(cur), boardID, threadID, replyID, "Reason 1")

	testing.SetRealm(testing.NewUserRealm(moderator))
	boards2.FlagReply(cross(cur), boardID, threadID, replyID, "Reason 2")
}

func main(cur realm) {
	flags := boards2.GetFlags(uint64(boardID), uint64(threadID), uint64(replyID), 0, 10)

	for _, f := range flags {
		println(f.User(), f.Reason())
	}
}

// Output:
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 Reason 1
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj Reason 2
