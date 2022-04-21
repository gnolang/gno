package boards

import (
	"std"
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
	std.AssertOriginCall()
	bid := incGetBoardID()
	caller := std.GetOrigCaller()
	url := "/r/boards:" + name
	board := newBoard(bid, url, name, caller)
	bidkey := boardIDKey(bid)
	gBoards.Set(bidkey, board)
	gBoardsByName.Set(name, board)
	return board.id
}

// TODO: rename to CreateThread.
func CreatePost(bid BoardID, title string, body string) PostID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	thread := board.AddThread(caller, title, body)
	return thread.id
}

func CreateReply(bid BoardID, threadid, postid PostID, body string) PostID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	thread := board.GetThread(threadid)
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
	board := getBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}
	dst := getBoard(dstBoardID)
	thread := board.GetThread(postid)
	repost := thread.AddRepostTo(caller, title, body, dst)
	return repost.id
}

func DeletePost(bid BoardID, threadid, postid PostID, reason string) {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	if postid == threadid {
		board.DeleteThread(threadid)
	} else {
		thread := board.GetThread(threadid)
		thread.DeletePost(postid)
	}
}
