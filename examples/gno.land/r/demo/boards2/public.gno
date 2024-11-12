package boards

import (
	"std"

	"gno.land/r/demo/users"
)

// TODO: Consider a simpler API initially by removing this type of functions
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

// TODO: Rename "postid" to "commentID", "threadid" to "threadID"
// TODO: Split into CreateComment(threadID) & CreateReply(commentID), check link generation functions first (ex. Post.GetReplyFromURL())
func CreateReply(bid BoardID, threadid, postid PostID, body string) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	board := mustGetBoard(bid)
	thread := board.GetThread(threadid) // TODO: Use found flag to encorage returned value validity checks
	if thread == nil {
		panic("thread not exist")
	}

	if postid == threadid {
		reply := thread.AddReply(caller, body)
		return reply.id
	} else {
		post := thread.GetReply(postid)
		reply := post.AddReply(caller, body)
		return reply.id
	}
}

// If dstBoard is private, does not ping back.
// If board specified by bid is private, panics.
func CreateRepost(bid BoardID, postid PostID, title string, body string, dstBoardID BoardID) PostID {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	assertAnonymousCallerFeeReceived(caller)

	board := mustGetBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}

	dst := mustGetBoard(dstBoardID)
	thread := board.GetThread(postid)
	if thread == nil {
		panic("thread not exist")
	}

	repost := thread.AddRepostTo(caller, title, body, dst)
	return repost.id
}

// TODO: Split into DeleteThread() & DeleteComment()
func DeletePost(bid BoardID, threadid, postid PostID, reason string) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread := board.GetThread(threadid)
	if thread == nil {
		panic("thread not exist")
	}

	if postid == threadid {
		// delete thread
		if !thread.HasPermission(caller, PermissionDelete) {
			panic("unauthorized")
		}
		board.DeleteThread(threadid)
	} else {
		// delete thread's post
		post := thread.GetReply(postid)
		if post == nil {
			panic("post not exist")
		}
		if !post.HasPermission(caller, PermissionDelete) {
			panic("unauthorized")
		}
		thread.DeletePost(postid)
	}
}

// TODO: Split into EditThread() & EditComment()
func EditPost(bid BoardID, threadid, postid PostID, title, body string) {
	assertIsUserCall()

	caller := std.GetOrigCaller()
	board := mustGetBoard(bid)
	thread := board.GetThread(threadid)
	if thread == nil {
		panic("thread not exist")
	}

	if postid == threadid {
		// edit thread
		if !thread.HasPermission(caller, PermissionEdit) {
			panic("unauthorized")
		}
		thread.Update(title, body)
	} else {
		// edit thread's post
		post := thread.GetReply(postid)
		if post == nil {
			panic("post not exist")
		}
		if !post.HasPermission(caller, PermissionEdit) {
			panic("unauthorized")
		}
		post.Update(title, body)
	}
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
