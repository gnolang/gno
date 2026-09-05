// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_flag_reply_06_filetest

package z_flag_reply_06_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	moderator address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid      boards.ID
	rid, tid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "body")

	// Invite a member using a role with permission to flag replies
	boards2.InviteMember(cross(cur), bid, moderator, boards2.RoleModerator)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(moderator))

	boards2.FlagReply(cross(cur), bid, tid, rid, "Reason")

	// Render content must contain a message about the hidden reply
	println(boards2.Render("test-board/1/2"))
}

// Output:
// ## Foo
//
// **[g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf)** `owner` · 2009-02-13 11:31pm UTC
//
//
// <gno-foreign>
// bar
// </gno-foreign>
//
//
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board/1/reply) [1] • [Edit](/r/gnoland/boards2/v1:test-board/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=2&threadID=1) • [Show all Replies](/r/gnoland/boards2/v1:test-board/1)
//
//
// >
// > **[g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf)** `owner` · 2009-02-13 11:31pm UTC [\#2](/r/gnoland/boards2/v1:test-board/1/2)
// > ⚠ Reply is hidden as it has been flagged as [inappropriate](/r/gnoland/boards2/v1:test-board/1/2/flagging-reasons)
