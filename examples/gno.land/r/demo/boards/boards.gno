package boards

import (
	"regexp"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/txlink"
)

//----------------------------------------
// Realm (package) state

var (
	gRealmPath      string      // Realm package path
	gBoards         avl.Tree    // id -> *Board
	gBoardsCtr      int         // increments Board.id
	gBoardsByName   avl.Tree    // name -> *Board
	gDefaultAnonFee = 100000000 // minimum fee required if anonymous
)

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)

func init() {
	gRealmPath = std.CurrentRealm().PkgPath()
}

func txlinkCall(fn string, args ...string) string {
	// Use gRealmPath to build TX link to allow links to be built during render calls
	return txlink.Realm(gRealmPath).Call(fn, args...)
}
