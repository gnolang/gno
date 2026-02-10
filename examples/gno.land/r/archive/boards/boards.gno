package boards

import (
	"chain/runtime"
	"regexp"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl"
)

//----------------------------------------
// Realm (package) state

var (
	gRealmLink      txlink.Realm
	gBoards         avl.Tree    // id -> *Board
	gBoardsCtr      int         // increments Board.id
	gBoardsByName   avl.Tree    // name -> *Board
	gDefaultAnonFee = 100000000 // minimum fee required if anonymous
)

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)

func init() {
	// Keep track of the realm package path to make it available on render
	gRealmLink = txlink.Realm(runtime.CurrentRealm().PkgPath())
}
