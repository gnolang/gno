package boards2

import (
	"std"
	"strings"

	"gno.land/r/demo/users"
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
	assertIsNotAnonymousCaller(caller)

	id := incGetBoardID()
	args := Args{name, id}
	// TODO: Do the callback really need the args or we could have the same result directly referencing?
	gAuth.WithPermission(caller, PermissionBoardCreate, args, func(a Args) { // TODO: Caller could be an extra func arg
		name := a[0].(string)
		id := a[1].(BoardID)
		board := newBoard(id, name, caller)
		gBoardsByID.Set(id.Key(), board)
		gBoardsByName.Set(name, board)
	})
	return id
}

func CreateThread(bid BoardID, title, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)
	assertBoardExists(bid)

	var (
		id   PostID // TODO: Decide how to deal with thread ID generation
		args = Args{bid, title, body}
	)

	gAuth.WithPermission(caller, PermissionThreadCreate, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)

		title := a[1].(string)
		body := a[2].(string)
		thread := board.AddThread(caller, title, body)
		id = thread.id // TODO: ID should be pre generated to allow async thread creation
	})
	return id
}

func CreateReply(bid BoardID, threadID, replyID PostID, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	board := mustGetBoard(bid)
	assertThreadExists(board, threadID)

	var (
		id   PostID // TODO: Decide how to deal with reply ID generation
		args = Args{bid, threadID, replyID, body}
	)

	gAuth.WithPermission(caller, PermissionReplyCreate, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)

		threadID := a[1].(PostID)
		thread := mustGetThread(board, threadID)

		var (
			reply   *Post
			replyID = a[2].(PostID)
			body    = a[3].(string)
		)

		if replyID == threadID {
			// When the parent reply is the thread just add reply to thread
			reply = thread.AddReply(caller, body)
		} else {
			// Try to get parent reply and add a new child reply
			post := mustGetReply(thread, replyID)
			reply = post.AddReply(caller, body)
		}

		id = reply.id // TODO: ID should be pre generated to allow async thread creation
	})
	return id
}

func CreateRepost(bid BoardID, threadID PostID, title, body string, dstBoardID BoardID) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)
	assertBoardExists(dstBoardID)

	board := mustGetBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}

	assertThreadExists(board, threadID)

	var (
		id   PostID // TODO: Decide how to deal with repost/thread ID generation
		args = Args{bid, dstBoardID, threadID, title, body}
	)

	gAuth.WithPermission(caller, PermissionThreadRepost, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)
		dstID := a[1].(BoardID)
		dst := mustGetBoard(dstID)

		threadID := a[2].(PostID)
		thread := mustGetThread(board, threadID)

		title := a[3].(string)
		body := a[4].(string)
		repost := thread.AddRepostTo(caller, title, body, dst)
		id = repost.id // TODO: ID should be pre generated to allow async thread creation
	})
	return id
}

func DeleteThread(bid BoardID, threadID PostID) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	assertThreadExists(board, threadID)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID}
	gAuth.WithPermission(caller, PermissionThreadDelete, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)

		threadID := a[1].(PostID)
		board.DeleteThread(threadID)
	})
}

func DeleteReply(bid BoardID, threadID, replyID PostID) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	assertReplyExists(thread, replyID)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID, replyID}
	gAuth.WithPermission(caller, PermissionReplyDelete, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)

		threadID := a[1].(PostID)
		thread := mustGetThread(board, threadID)

		replyID := a[2].(PostID)
		thread.DeleteReply(replyID)
	})
}

func EditThread(bid BoardID, threadID PostID, title, body string) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	assertThreadExists(board, threadID)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID, title, body}
	gAuth.WithPermission(caller, PermissionThreadEdit, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)

		threadID := a[1].(PostID)
		thread := mustGetThread(board, threadID)

		title := a[2].(string)
		body := a[3].(string)
		thread.Update(title, body)
	})
}

func EditReply(bid BoardID, threadID, replyID PostID, title, body string) {
	assertIsUserCall()

	board := mustGetBoard(bid)
	thread := mustGetThread(board, threadID)
	assertReplyExists(thread, replyID)

	caller := std.GetOrigCaller()
	args := Args{bid, threadID, replyID, title, body}
	gAuth.WithPermission(caller, PermissionReplyEdit, args, func(a Args) {
		bid := a[0].(BoardID)
		board := mustGetBoard(bid)

		threadID := a[1].(PostID)
		thread := mustGetThread(board, threadID)

		replyID := a[2].(PostID)
		reply := mustGetReply(thread, replyID)

		title := a[3].(string)
		body := a[4].(string)
		reply.Update(title, body)
	})
}

func assertIsUserCall() {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
}

func assertIsNotAnonymousCaller(caller std.Address) {
	// Caller is anonymous if doesn't have a registered user name
	if users.GetUserByAddress(caller) == nil {
		panic("unauthorized")
	}
}

func assertAnonymousFeeReceived() {
	sent := std.GetOrigSend()
	fee := std.NewCoin("ugnot", int64(defaultAnonymousFee))
	if len(sent) == 0 || sent[0].IsLT(fee) {
		panic("please register a user, otherwise a minimum fee of " + fee.String() + " is required")
	}
	return
}

func assertAnonymousCallerFeeReceived(caller std.Address) {
	if users.GetUserByAddress(caller) == nil {
		assertAnonymousFeeReceived()
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
