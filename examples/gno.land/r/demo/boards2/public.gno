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
	if name == "" {
		panic("board name is empty")
	}

	caller := std.GetOrigCaller()
	id := incGetBoardID()
	args := Args{name, id}
	gPerm.WithPermission(caller, PermissionBoardCreate, args, func(Args) {
		if gBoardsByName.Has(name) {
			panic("board already exists")
		}

		board := newBoard(id, name, caller)
		gBoardsByID.Set(id.Key(), board)
		gBoardsByName.Set(name, board)
	})
	return id
}

func FlagThread(bid BoardID, postID PostID, reason string) {
	caller := std.GetOrigCaller()
	// TODO: Replace by assertHasBoardPermission() when supported
	assertHasPermission(caller, PermissionThreadFlag)

	board := mustGetBoard(bid)
	t, ok := board.GetThread(postID)
	if !ok {
		panic("post doesn't exist")
	}

	if flagItem(t, NewFlag(caller, reason)) {
		t.SetVisible(false)
	}
}

func CreateThread(bid BoardID, title, body string) PostID {
	assertIsUserCall()

	// TODO: Assert that caller is a board member (when board type is invite only)
	caller := std.GetOrigCaller()
	assertHasPermission(caller, PermissionThreadCreate) // TODO: Who can create threads?
	assertBoardExists(bid)

	board := mustGetBoard(bid)
	thread := board.AddThread(caller, title, body)
	return thread.id
}

func CreateReply(bid BoardID, threadID, replyID PostID, body string) PostID {
	assertIsUserCall()

	// TODO: Assert that caller is a board member (when board type is invite only)
	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)

	assertThreadVisible(thread)

	// TODO: Assert thread is not locked
	// TODO: Assert that caller is a board member (when board type is invite only)
	var reply *Post
	if replyID == threadID {
		// When the parent reply is the thread just add reply to thread
		reply = thread.AddReply(caller, body)
	} else {
		// Try to get parent reply and add a new child reply
		post := mustGetReply(thread, replyID)
		assertThreadVisible(thread)

		reply = post.AddReply(caller, body)
	}
	return reply.id
}

func FlagReply(bid BoardID, threadID, replyID PostID, reason string) {
	caller := std.GetOrigCaller()
	// TODO: Replace by assertHasBoardPermission() when supported
	assertHasPermission(caller, PermissionThreadFlag)

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)

	if flagItem(reply, NewFlag(caller, reason)) {
		reply.SetVisible(false)
	}
}

func CreateRepost(bid BoardID, threadID PostID, title, body string, dstBoardID BoardID) PostID {
	assertIsUserCall()

	// TODO: Assert that caller is a board member (when board type is invite only)
	caller := std.GetOrigCaller()
	assertBoardExists(dstBoardID)

	board := mustGetBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}

	// TODO: Assert that board allows reposts?
	// TODO: Assert that caller is member of both boards (when board types are invite only)

	assertThreadExists(board, threadID)

	dst := mustGetBoard(dstBoardID)
	thread := mustGetThread(board, threadID)
	repost := thread.AddRepostTo(caller, title, body, dst)
	return repost.id
}

func DeleteThread(bid BoardID, threadID PostID) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	assertThreadExists(board, threadID)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID}
	gPerm.WithPermission(caller, PermissionThreadDelete, args, func(Args) {
		board := mustGetBoard(bid)
		board.DeleteThread(threadID)
	})
}

func DeleteReply(bid BoardID, threadID, replyID PostID) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	assertReplyExists(thread, replyID)

	// TODO: Hide reply when the caller is the owner of the reply (remove WithPermission call for now)
	// TODO: Support removing reply and children though proposals? (WithPermission)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID, replyID}
	gPerm.WithPermission(caller, PermissionReplyDelete, args, func(a Args) {
		board := mustGetBoard(bid)
		thread := mustGetThread(board, threadID)
		thread.DeleteReply(replyID)
	})
}

func EditThread(bid BoardID, threadID PostID, title, body string) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	assertThreadExists(board, threadID)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID, title, body}
	gPerm.WithPermission(caller, PermissionThreadEdit, args, func(Args) {
		board := mustGetBoard(bid)
		thread := mustGetThread(board, threadID)
		thread.Update(title, body)
	})
}

func EditReply(bid BoardID, threadID, replyID PostID, title, body string) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	caller := std.GetOrigCaller()
	if caller != reply.GetCreator() {
		panic("only the reply creator is allowed to edit it")
	}

	// TODO: Should we have a history of previous reply contents?
	reply.Update(title, body)
}

func InviteMember(user std.Address, role Role) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	args := Args{user, role}
	gPerm.WithPermission(caller, PermissionMemberInvite, args, func(Args) {
		if err := gPerm.AddUser(user, role); err != nil {
			panic(err)
		}
	})
}

func RemoveMember(user std.Address) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	gPerm.WithPermission(caller, PermissionMemberRemove, Args{user}, func(Args) {
		if !gPerm.RemoveUser(user) {
			panic("member not found")
		}
	})
}

func assertIsUserCall() {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
}

func assertHasPermission(user std.Address, p Permission) {
	if !gPerm.HasPermission(user, p) {
		panic("unauthorized")
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
	if thread.Hidden() {
		panic("thread with ID: " + thread.GetPostID().String() + " was hidden")
	}
}
