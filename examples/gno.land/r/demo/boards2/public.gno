package boards

import (
	"std"
	"strconv"
)

//----------------------------------------
// Public facing functions

// TODO: Consider a simpler API initially by removing this type of functions
func GetBoardIDFromName(name string) (BoardID, bool) {
	boardI, exists := gBoardsByName.Get(name)
	if !exists {
		return 0, false
	}
	return boardI.(*Board).id, true
}

func CreateBoard(name string) BoardID {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
	bid := incGetBoardID()
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		panic("unauthorized")
	}
	url := "/r/demo/boards:" + name
	board := newBoard(bid, url, name, caller)
	gBoards.Set(bid.Key(), board)
	gBoardsByName.Set(name, board)
	return board.id
}

// TODO: This must be an assertion function
// TODO: Should we display realm's amount? Transfer to a pool realm/DAO?
func checkAnonFee() bool {
	sent := std.GetOrigSend()
	anonFeeCoin := std.NewCoin("ugnot", int64(gDefaultAnonFee))
	if len(sent) == 1 && sent[0].IsGTE(anonFeeCoin) {
		return true
	}
	return false
}

func CreateThread(bid BoardID, title string, body string) PostID {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		if !checkAnonFee() {
			panic("please register, otherwise minimum fee " + strconv.Itoa(gDefaultAnonFee) + " is required if anonymous")
		}
	}
	board := getBoard(bid)
	if board == nil {
		panic("board not exist")
	}
	thread := board.AddThread(caller, title, body)
	return thread.id
}

// TODO: Rename "postid" to "commentID", "threadid" to "threadID"
// TODO: Split into CreateComment(threadID) & CreateReply(commentID), check link generation functions first (ex. Post.GetReplyFromURL())
func CreateReply(bid BoardID, threadid, postid PostID, body string) PostID {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) { // TODO: Create assertIsUser() function
		panic("invalid non-user call")
	}
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		if !checkAnonFee() {
			panic("please register, otherwise minimum fee " + strconv.Itoa(gDefaultAnonFee) + " is required if anonymous")
		}
	}
	board := getBoard(bid) // TODO: Create a mustGetBoard() function
	if board == nil {
		panic("board not exist")
	}
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
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		// TODO: allow with gDefaultAnonFee payment.
		if !checkAnonFee() {
			panic("please register, otherwise minimum fee " + strconv.Itoa(gDefaultAnonFee) + " is required if anonymous")
		}
	}
	board := getBoard(bid)
	if board == nil {
		panic("src board not exist")
	}
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}
	dst := getBoard(dstBoardID)
	if dst == nil {
		panic("dst board not exist")
	}
	thread := board.GetThread(postid)
	if thread == nil {
		panic("thread not exist")
	}
	repost := thread.AddRepostTo(caller, title, body, dst)
	return repost.id
}

// TODO: Split into DeleteThread() & DeleteComment()
func DeletePost(bid BoardID, threadid, postid PostID, reason string) {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	if board == nil {
		panic("board not exist")
	}
	thread := board.GetThread(threadid)
	if thread == nil {
		panic("thread not exist")
	}
	if postid == threadid {
		// delete thread
		if !thread.HasPermission(caller, DeletePermission) {
			panic("unauthorized")
		}
		board.DeleteThread(threadid)
	} else {
		// delete thread's post
		post := thread.GetReply(postid)
		if post == nil {
			panic("post not exist")
		}
		if !post.HasPermission(caller, DeletePermission) {
			panic("unauthorized")
		}
		thread.DeletePost(postid)
	}
}

// TODO: Split into EditThread() & EditComment()
func EditPost(bid BoardID, threadid, postid PostID, title, body string) {
	if !(std.IsOriginCall() || std.PrevRealm().IsUser()) {
		panic("invalid non-user call")
	}
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	if board == nil {
		panic("board not exist")
	}
	thread := board.GetThread(threadid)
	if thread == nil {
		panic("thread not exist")
	}
	if postid == threadid {
		// edit thread
		if !thread.HasPermission(caller, EditPermission) {
			panic("unauthorized")
		}
		thread.Update(title, body)
	} else {
		// edit thread's post
		post := thread.GetReply(postid)
		if post == nil {
			panic("post not exist")
		}
		if !post.HasPermission(caller, EditPermission) {
			panic("unauthorized")
		}
		post.Update(title, body)
	}
}
