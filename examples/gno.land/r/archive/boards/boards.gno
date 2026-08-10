package boards

import (
	"regexp"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl/v0"
)

//----------------------------------------
// Realm (package) state

var (
	gRealmLink      = txlink.Realm("gno.land/r/archive/boards")
	gBoards         avl.Tree    // id -> *Board
	gBoardsCtr      int         // increments Board.id
	gBoardsByName   avl.Tree    // name -> *Board
	gDefaultAnonFee = 100000000 // minimum fee required if anonymous
)

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)
