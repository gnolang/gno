package boards2

import "gno.land/p/demo/avl"

// Default minimum fee in ugnot required for anonymous users
const defaultAnonymousFee = 100_000_000

var (
	gLastBoardID  BoardID
	gBoardsByID   avl.Tree // string(id) -> *Board
	gBoardsByName avl.Tree // string(name) -> *Board
)

// incGetBoardID returns a new board ID.
func incGetBoardID() BoardID {
	gLastBoardID++
	return gLastBoardID
}

// getBoard returns a board for a specific ID.
func getBoard(id BoardID) (_ *Board, found bool) {
	v, exists := gBoardsByID.Get(id.Key())
	if !exists {
		return nil, false
	}
	return v.(*Board), true
}

// mustGetBoard returns a board or panics when it's not found.
func mustGetBoard(id BoardID) *Board {
	board, found := getBoard(id)
	if !found {
		panic("board does not exist with ID: " + id.String())
	}
	return board
}

// mustGetThread returns a thread or panics when it's not found.
func mustGetThread(board *Board, threadID PostID) *Post {
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread does not exist with ID: " + threadID.String())
	}
	return thread
}

// mustGetReply returns a reply or panics when it's not found.
func mustGetReply(thread *Post, replyID PostID) *Post {
	reply, found := thread.GetReply(replyID)
	if !found {
		panic("reply does not exist with ID: " + replyID.String())
	}
	return reply
}
