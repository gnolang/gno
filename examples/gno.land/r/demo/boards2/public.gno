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

	// TODO: This and other cases must not be async (though proposals for example)
	// TODO: Extend the permissioner to support cases like this one (`IsAllowed()` method?)
	// NOTE: Args are "tight" to the permission, maybe we can join them into a permission struct/interface
	var id BoardID
	gAuth.WithPermission(caller, PermissionBoardCreate, nil, func(a Args) {
		id := incGetBoardID()
		board := newBoard(id, name, caller)
		gBoardsByID.Set(id.Key(), board)
		gBoardsByName.Set(name, board)
		id = board.id
	})
	return id
}

func CreateThread(bid BoardID, title string, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	var (
		id    PostID
		board = mustGetBoard(bid)
	)

	gAuth.WithPermission(caller, PermissionThreadCreate, nil, func(a Args) {
		thread := board.AddThread(caller, title, body)
		id = thread.id
	})
	return id
}

func CreateReply(bid BoardID, threadID, replyID PostID, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	var (
		reply  *Post
		board  = mustGetBoard(bid)
		thread = mustGetThread(board, threadID)
	)

	gAuth.WithPermission(caller, PermissionReplyCreate, nil, func(a Args) {
		if replyID == threadID {
			// When the parent reply is the thread just add reply to thread
			reply = thread.AddReply(caller, body)
		} else {
			// Try to get parent reply and add a new child reply
			post := mustGetReply(thread, replyID)
			reply = post.AddReply(caller, body)
		}
	})
	return reply.id
}

func CreateRepost(bid BoardID, threadID PostID, title string, body string, dstBoardID BoardID) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	board := mustGetBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}

	var (
		id     PostID
		dst    = mustGetBoard(dstBoardID)
		thread = mustGetThread(board, threadID)
	)

	gAuth.WithPermission(caller, PermissionThreadRepost, nil, func(a Args) {
		repost := thread.AddRepostTo(caller, title, body, dst)
		id = repost.id
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
