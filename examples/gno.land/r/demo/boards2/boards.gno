package boards

import (
	"gno.land/p/demo/avl"
)

// TODO: Create a Boards or App struct to handle counter and tree indexes
var (
	gBoards         avl.Tree    // id -> *Board
	gBoardsCtr      int         // increments Board.id
	gBoardsByName   avl.Tree    // name -> *Board
	gDefaultAnonFee = 100000000 // minimum fee required if anonymous
)
