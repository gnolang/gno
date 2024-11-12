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

func getBoard(bid BoardID) *Board {
	v, exists := gBoards.Get(bid.Key())
	if !exists {
		return nil
	}
	return v.(*Board)
}

func incGetBoardID() BoardID {
	gBoardsCtr++
	return BoardID(gBoardsCtr)
}
