package boards2

import (
	"std"
	"strings"
)

func GetBoardIDFromName(name string) (BoardID, bool) {
	v, found := gBoardsByName.Get(name)
	if !found {
		return 0, false
	}
	return v.(*Board).id, true
}

func CreateBoard(name string) BoardID {
	assertIsUserCall()

	name = strings.TrimSpace(name)
	assertNameIsNotEmpty(name)
	assertBoardNameNotExists(name)

	caller := std.GetOrigCaller()
	id := incGetBoardID()
	args := Args{name, id}
	gPerm.WithPermission(caller, PermissionBoardCreate, args, func(Args) {
		assertBoardNameNotExists(name)

		board := newBoard(id, name, caller)
		gBoardsByID.Set(id.Key(), board)
		gBoardsByName.Set(name, board)
	})
	return id
}

func RenameBoard(name, newName string) {
	assertIsUserCall()

	newName = strings.TrimSpace(newName)
	assertNameIsNotEmpty(newName)
	assertBoardNameNotExists(newName)

	board := mustGetBoardByName(name)
	bid := board.GetID()
	caller := std.GetOrigCaller()
	args := Args{bid, name, newName}
	board.perms.WithPermission(caller, PermissionBoardRename, args, func(Args) {
		assertBoardNameNotExists(newName)

		board := mustGetBoard(bid)
		board.aliases = append(board.aliases, board.name)
		board.name = newName

		// Index board for the new name keeping previous indexes for older names
		gBoardsByName.Set(newName, board)
	})
}

func FlagThread(bid BoardID, postID PostID, reason string) {
	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	assertHasBoardPermission(board, caller, PermissionThreadFlag)

	t, ok := board.GetThread(postID)
	if !ok {
		panic("post doesn't exist")
	}

	hide := flagItem(t, NewFlag(caller, reason))
	if hide {
		t.SetVisible(false)
	}
}

func CreateThread(bid BoardID, title, body string) PostID {
	assertIsUserCall()

	title = strings.TrimSpace(title)
	assertTitleIsNotEmpty(title)

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	assertHasBoardPermission(board, caller, PermissionThreadCreate)

	thread := board.AddThread(caller, title, body)
	return thread.id
}

func CreateReply(bid BoardID, threadID, replyID PostID, body string) PostID {
	assertIsUserCall()

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	assertHasBoardPermission(board, caller, PermissionReplyCreate)

	thread := mustGetThread(board, threadID)
	assertThreadVisible(thread)

	var reply *Post
	if replyID == 0 {
		// When the parent reply is the thread just add reply to thread
		reply = thread.AddReply(caller, body)
	} else {
		// Try to get parent reply and add a new child reply
		post := mustGetReply(thread, replyID)
		assertReplyVisible(post)

		reply = post.AddReply(caller, body)
	}
	return reply.id
}

func FlagReply(bid BoardID, threadID, replyID PostID, reason string) {
	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	assertHasBoardPermission(board, caller, PermissionThreadFlag)

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)

	hide := flagItem(reply, NewFlag(caller, reason))
	if hide {
		reply.SetVisible(false)
	}
}

func CreateRepost(bid BoardID, threadID PostID, title, body string, dstBoardID BoardID) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	dst := mustGetBoard(dstBoardID)
	assertHasBoardPermission(dst, caller, PermissionThreadRepost)

	board := mustGetBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}

	thread := mustGetThread(board, threadID)
	repostId := dst.incGetPostID()
	repost := newPost(dst, repostId, caller, title, body, repostId, thread.GetPostID(), thread.GetBoard().GetID())
	thread.AddRepostTo(caller, repost, dst)
	return repostId
}

func DeleteThread(bid BoardID, threadID PostID) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	if caller != thread.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionThreadDelete)
	}

	// TODO: Discuss how to deal with thread deletion
	board.DeleteThread(threadID)
}

func DeleteReply(bid BoardID, threadID, replyID PostID) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	assertReplyVisible(reply)

	caller := std.GetOrigCaller()
	if caller != reply.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionReplyDelete)
	}

	// Soft delete reply by changing its body when it contains
	// sub-replies, otherwise hard delete it.
	if reply.HasReplies() {
		reply.Update(reply.GetTitle(), "This reply has been deleted")
	} else {
		thread.DeleteReply(replyID)
	}
}

func EditThread(bid BoardID, threadID PostID, title, body string) {
	assertIsUserCall()

	title = strings.TrimSpace(title)
	assertTitleIsNotEmpty(title)

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	caller := std.GetOrigCaller()
	if caller != thread.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionThreadEdit)
	}

	thread.Update(title, body)
}

func EditReply(bid BoardID, threadID, replyID PostID, body string) {
	assertIsUserCall()

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	assertReplyVisible(reply)

	if std.GetOrigCaller() != reply.GetCreator() {
		panic("only the reply creator is allowed to edit it")
	}

	reply.Update("", body)
}

func InviteMember(bid BoardID, user std.Address, role Role) {
	assertIsUserCall()

	perms := mustGetPermissions(bid)
	caller := std.GetOrigCaller()
	args := Args{user, role}
	perms.WithPermission(caller, PermissionMemberInvite, args, func(Args) {
		if err := perms.AddUser(user, role); err != nil {
			panic(err)
		}
	})
}

func RemoveMember(bid BoardID, user std.Address) {
	assertIsUserCall()

	perms := mustGetPermissions(bid)
	caller := std.GetOrigCaller()
	perms.WithPermission(caller, PermissionMemberRemove, Args{user}, func(Args) {
		if !perms.RemoveUser(user) {
			panic("member not found")
		}
	})
}

func IsMember(bid BoardID, member std.Address) bool {
	perms := mustGetPermissions(bid)
	return perms.HasUser(member)
}

func HasMemberRole(bid BoardID, member std.Address, role Role) bool {
	perms := mustGetPermissions(bid)
	return perms.HasRole(member, role)
}

func ChangeMemberRole(bid BoardID, member std.Address, role Role) {
	assertIsUserCall()

	perms := mustGetPermissions(bid)
	caller := std.GetOrigCaller()
	args := Args{bid, member, role}
	perms.WithPermission(caller, PermissionRoleChange, args, func(Args) {
		if err := perms.SetUserRoles(member, role); err != nil {
			panic(err)
		}
	})
}

func assertIsUserCall() {
	if !std.PrevRealm().IsUser() {
		panic("invalid non-user call")
	}
}

func assertHasBoardPermission(b *Board, user std.Address, p Permission) {
	if !b.perms.HasPermission(user, p) {
		panic("unauthorized")
	}
}

func assertBoardExists(id BoardID) {
	if _, found := getBoard(id); !found {
		panic("board not found: " + id.String())
	}
}

func assertNameIsNotEmpty(name string) {
	if name == "" {
		panic("name is empty")
	}
}

func assertTitleIsNotEmpty(title string) {
	if title == "" {
		panic("title is empty")
	}
}

func assertBodyIsNotEmpty(body string) {
	if body == "" {
		panic("body is empty")
	}
}

func assertBoardNameNotExists(name string) {
	if gBoardsByName.Has(name) {
		panic("board already exists")
	}
}

func assertThreadExists(b *Board, threadID PostID) {
	if _, found := b.GetThread(threadID); !found {
		panic("thread not found: " + threadID.String())
	}
}

func assertReplyExists(thread *Post, replyID PostID) {
	if _, found := thread.GetReply(replyID); !found {
		panic("reply not found: " + replyID.String())
	}
}

func assertThreadVisible(thread *Post) {
	if thread.IsHidden() {
		panic("thread with ID: " + thread.GetPostID().String() + " was hidden")
	}
}

func assertReplyVisible(thread *Post) {
	if thread.IsHidden() {
		panic("reply with ID: " + thread.GetPostID().String() + " was hidden")
	}
}

func mustGetPermissions(bid BoardID) Permissions {
	if bid != 0 {
		board := mustGetBoard(bid)
		return board.perms
	}
	return gPerm
}
