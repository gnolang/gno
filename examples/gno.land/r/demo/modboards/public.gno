package boards

import (
	"std"
	"strconv"

	"gno.land/p/demo/flags_index"
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

func CreateBoard(name string) BoardID {
	bid := incGetBoardID()
	caller := std.PrevRealm().Addr()
	url := "/r/demo/modboards:" + name
	board := newBoard(bid, url, name, caller)
	bidkey := boardIDKey(bid)
	gBoards.Set(bidkey, board)
	gBoardsByName.Set(name, board)
	return board.id
}

func checkAnonFee() bool {
	sent := std.GetOrigSend()
	anonFeeCoin := std.Coin{"ugnot", int64(gDefaultAnonFee)}
	if len(sent) == 1 && sent[0].IsGTE(anonFeeCoin) {
		return true
	}
	return false
}

func CreateThread(bid BoardID, title string, body string) PostID {
	std.AssertOriginCall()
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

func CreateReply(bid BoardID, threadid, postid PostID, body string) PostID {
	std.AssertOriginCall()
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
func CreateRepost(bid BoardID, postid PostID, title string, body string, dstBoardID BoardID) PostID {
	std.AssertOriginCall()
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

func DeletePost(bid BoardID, threadid, postid PostID, reason string) {
	caller := std.PrevRealm().Addr()
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
	board.flags.ClearFlagCount(getFlagID(threadid, postid))
}

func EditPost(bid BoardID, threadid, postid PostID, title, body string) {
	std.AssertOriginCall()
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

// Maybe we should require a deposit that is returned if the post is deleted
func FlagPost(boardID BoardID, threadID PostID, postID PostID) {
	// check that the target exists
	board := getBoard(boardID)
	if board == nil {
		panic("board not exist")
	}
	thread := board.GetThread(threadID)
	if thread == nil {
		panic("thread not exist")
	}
	if postID != threadID {
		post := thread.GetReply(postID)
		if post == nil {
			panic("post not exist")
		}
	}

	// flag
	board.flags.Flag(getFlagID(threadID, postID), std.PrevRealm().Addr().String())
}
