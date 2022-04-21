package boards

import (
	"regexp"

	"gno.land/p/avl"
)

//----------------------------------------
// Realm (package) state

var gBoards *avl.Tree       // id -> *Board
var gBoardsCtr int          // increments Board.id
var gBoardsByName *avl.Tree // name -> *Board

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)
