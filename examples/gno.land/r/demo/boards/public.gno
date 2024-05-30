package boards

import (
	"std"
	"strconv"

	"gno.land/p/demo/boards"
)

//----------------------------------------
// Public facing functions

func GetBoardIDFromName(name string) (boards.BoardID, bool) {
	boardI, exists := gBoardsByName.Get(name)
	if !exists {
		return 0, false
	}
	return boardI.(*boards.Board).Id(), true
}

func CreateBoard(name string) boards.BoardID {
	std.AssertOriginCall()
	bid := incGetBoardID()
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		panic("unauthorized")
	}
	url := "/r/demo/boards:" + name
	if !reName.MatchString(name) {
		panic("invalid name: " + name)
	}
	exists := gBoardsByName.Has(name)
	if exists {
		panic("board already exists")
	}
	board := boards.NewBoard(bid, url, name, caller)
	bidkey := boards.BoardIDKey(bid)
	gBoards.Set(bidkey, board)
	gBoardsByName.Set(name, board)
	return board.Id()
}

func checkAnonFee() bool {
	sent := std.GetOrigSend()
	anonFeeCoin := std.NewCoin("ugnot", int64(gDefaultAnonFee))
	if len(sent) == 1 && sent[0].IsGTE(anonFeeCoin) {
		return true
	}
	return false
}

func CreateThread(bid boards.BoardID, title string, body string) boards.PostID {
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
	return thread.Id()
}

func CreateReply(bid boards.BoardID, threadid, postid boards.PostID, body string) boards.PostID {
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
		return reply.Id()
	} else {
		post := thread.GetReply(postid)
		reply := post.AddReply(caller, body)
		return reply.Id()
	}
}

// If dstBoard is private, does not ping back.
// If board specified by bid is private, panics.
func CreateRepost(bid boards.BoardID, postid boards.PostID, title string, body string, dstBoardID boards.BoardID) boards.PostID {
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
	return repost.Id()
}

func DeletePost(bid boards.BoardID, threadid, postid boards.PostID, reason string) {
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
		// delete thread
		if !thread.HasPermission(caller, boards.DeletePermission) {
			panic("unauthorized")
		}
		board.DeleteThread(threadid)
	} else {
		// delete thread's post
		post := thread.GetReply(postid)
		if post == nil {
			panic("post not exist")
		}
		if !post.HasPermission(caller, boards.DeletePermission) {
			panic("unauthorized")
		}
		thread.DeletePost(postid)
	}
}

func EditPost(bid boards.BoardID, threadid, postid boards.PostID, title, body string) {
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
		if !thread.HasPermission(caller, boards.EditPermission) {
			panic("unauthorized")
		}
		thread.Update(title, body)
	} else {
		// edit thread's post
		post := thread.GetReply(postid)
		if post == nil {
			panic("post not exist")
		}
		if !post.HasPermission(caller, boards.EditPermission) {
			panic("unauthorized")
		}
		post.Update(title, body)
	}
}
