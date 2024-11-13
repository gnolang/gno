package boards

import "gno.land/p/demo/avl"

// Default minimum fee in ugnot required for anonymous users
const defaultAnonymousFee = 100_000_000

// TODO: Create a Boards or App struct to handle counter and tree indexes
var (
	gBoards       avl.Tree // id -> *Board
	gBoardsCtr    int      // increments Board.id
	gBoardsByName avl.Tree // name -> *Board
)

// getBoard returns a boards for a specific ID.
func getBoard(id BoardID) (_ *Board, found bool) {
	v, exists := gBoards.Get(id.Key())
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
	reply := thread.GetReply(replyID)
	if reply == nil {
		panic("reply does not exist with ID: " + replyID.String())
	}
	return reply
}

func incGetBoardID() BoardID {
	gBoardsCtr++
	return BoardID(gBoardsCtr)
}
