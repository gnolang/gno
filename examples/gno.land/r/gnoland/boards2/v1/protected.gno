// The `protected.gno` file contains public realm functions that must only be called
// from realms that live within the Boards2 package namespace. This allows sub realms
// to access boards data to be able to migrate content from one version to another or
// to implement specific features in separate sub realms.
package boards2

import (
	"chain/runtime"
	"path"
	"strings"

	"gno.land/p/gnoland/boards"
)

// Keeps the package path to Boards2 realm.
// It ignores the last realm path element which is the current version.
var boardsNS = path.Dir(runtime.CurrentRealm().PkgPath()) + "/"

// GetBoard returns a board.
// This is a protected function only callable by Boards2 sub realms.
func GetBoard(boardID boards.ID) (_ *boards.Board, found bool) {
	assertRealmHasBoardsNS(runtime.CurrentRealm())
	// NOTE: Should we also check that origin caller is currently an owner within the board?
	return gBoards.Get(boardID)
}

// MustGetBoard returns a board or panics on error.
// This is a protected function only callable by Boards2 sub realms.
func MustGetBoard(boardID boards.ID) *boards.Board {
	assertRealmHasBoardsNS(runtime.CurrentRealm())
	// NOTE: Should we also check that origin caller is currently an owner within the board?
	return mustGetBoard(boardID)
}

// assertRealmHasBoardsNS asserts that a realm lives within Boards2 namespace.
func assertRealmHasBoardsNS(r runtime.Realm) {
	if !strings.HasPrefix(r.PkgPath(), boardsNS) {
		panic("forbidden")
	}
}
