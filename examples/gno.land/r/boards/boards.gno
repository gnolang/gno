package boards

import (
	"regexp"

	"gno.land/p/avl"
)

//----------------------------------------
// Realm (package) state

var gBoards = avl.NewMutTree()       // id -> *Board
var gBoardsCtr int                   // increments Board.id
var gBoardsByName = avl.NewMutTree() // name -> *Board

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)
