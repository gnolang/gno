package boards

import (
	"regexp"

	"gno.land/p/demo/avl"
)

//----------------------------------------
// Realm (package) state

var (
	gBoards         = avl.Tree // id -> *Board
	gBoardsCtr      int                // increments Board.id
	gBoardsByName   = avl.Tree // name -> *Board
	gDefaultAnonFee = 100000000        // minimum fee required if anonymous
)

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)
