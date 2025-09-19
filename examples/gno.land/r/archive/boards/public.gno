package boards

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strconv"
)

//----------------------------------------
// Public facing functions

func GetBoardIDFromName(name string) (BoardID, bool) {
	boardI, exists := gBoardsByName.Get(name)
	if !exists {
		return 0, false
	}
	return boardI.(*Board).id, true
}

func CreateBoard(cur realm, name string) BoardID {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-user call")
	}

	bid := incGetBoardID()
	caller := runtime.OriginCaller()
	if usernameOf(caller) == "" {
		panic("unauthorized")
	}
	url := "/r/archive/boards:" + name
	board := newBoard(bid, url, name, caller)
	bidkey := boardIDKey(bid)
	gBoards.Set(bidkey, board)
	gBoardsByName.Set(name, board)
	return board.id
}

func checkAnonFee() bool {
	sent := banker.OriginSend()
	anonFeeCoin := chain.NewCoin("ugnot", int64(gDefaultAnonFee))
	if len(sent) == 1 && sent[0].IsGTE(anonFeeCoin) {
		return true
	}
	return false
}

func CreateThread(cur realm, bid BoardID, title string, body string) PostID {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-user call")
	}

	caller := runtime.OriginCaller()
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

func CreateReply(cur realm, bid BoardID, threadid, postid PostID, body string) PostID {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-user call")
	}

	caller := runtime.OriginCaller()
	if usernameOf(caller) == "" {
		if !checkAnonFee() {
			panic("please register, otherwise minimum fee " + strconv.Itoa(gDefaultAnonFee) + " is required if anonymous")
		}
	}
	board := getBoard(bid)
	if board == nil {
		panic("board not exist")
	}
	thread := board.GetThread(threadid)
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
func CreateRepost(cur realm, bid BoardID, postid PostID, title string, body string, dstBoardID BoardID) PostID {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-user call")
	}

	caller := runtime.OriginCaller()
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

func DeletePost(cur realm, bid BoardID, threadid, postid PostID, reason string) {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-user call")
	}

	caller := runtime.OriginCaller()
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

func EditPost(cur realm, bid BoardID, threadid, postid PostID, title, body string) {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-user call")
	}

	caller := runtime.OriginCaller()
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
