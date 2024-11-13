package boards

import (
	"std"

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

	caller := std.GetOrigCaller()
	assertIsNotAnonymousCaller(caller)

	bid := incGetBoardID()
	url := "/r/demo/boards:" + name
	board := newBoard(bid, url, name, caller)
	gBoards.Set(bid.Key(), board)
	gBoardsByName.Set(name, board)
	return board.id
}

func CreateThread(bid BoardID, title string, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	board := mustGetBoard(bid)
	thread := board.AddThread(caller, title, body)
	return thread.id
}

func CreateReply(bid BoardID, threadID, replyID PostID, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	board := mustGetBoard(bid)
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread not exist")
	}

	if replyID == threadID {
		reply := thread.AddReply(caller, body)
		return reply.id
	}

	post := thread.GetReply(replyID)
	reply := post.AddReply(caller, body)
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

	dst := mustGetBoard(dstBoardID)
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread not exist")
	}

	repost := thread.AddRepostTo(caller, title, body, dst)
	return repost.id
}

func DeleteThread(bid BoardID, threadID PostID, reason string) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread not found")
	}

	if !thread.HasPermission(caller, PermissionDelete) {
		panic("unauthorized")
	}
	board.DeleteThread(threadID)
}

func DeleteReply(bid BoardID, threadID, replyID PostID, reason string) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread, found := board.GetThread(threadID) // TODO: Add mustGetThread(board, threadID)
	if !found {
		panic("thread not found")
	}

	reply := thread.GetReply(replyID) // TODO: Add mustGetReply(thread, replyID)
	if reply == nil {
		panic("reply not found")
	}
	if !reply.HasPermission(caller, PermissionDelete) { // TODO: Add AssertHasPermission()
		panic("unauthorized")
	}
	thread.DeleteReply(replyID)
}

func EditThread(bid BoardID, threadID PostID, title, body string) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread not found")
	}

	if !thread.HasPermission(caller, PermissionEdit) {
		panic("unauthorized")
	}
	thread.Update(title, body)
}

func EditReply(bid BoardID, threadID, replyID PostID, title, body string) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread not exist")
	}

	post := thread.GetReply(replyID)
	if post == nil {
		panic("post not found")
	}
	if !post.HasPermission(caller, PermissionEdit) {
		panic("unauthorized")
	}
	post.Update(title, body)
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
	// TODO: Should we display realm's amount somewhere or transfer to a pool realm/DAO?
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
