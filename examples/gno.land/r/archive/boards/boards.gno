package boards

import (
	"chain/runtime"
	"regexp"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl/v0"
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
	// SECURITY XXX: this captures runtime.CurrentRealm().PkgPath() at
	// init time and stores it in a package var. PkgPath is stable per
	// realm so this is safe today, but storing realm-derived values in
	// package state is the same anti-pattern as storing realm values
	// themselves — easy to mistake for "OK to do" elsewhere. Fix: just
	// hardcode the realm pkgPath as a const string; no init-time runtime
	// walk needed.
	gRealmLink = txlink.Realm(runtime.CurrentRealm().PkgPath())
}
