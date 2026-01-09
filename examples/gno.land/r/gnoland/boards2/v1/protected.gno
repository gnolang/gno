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

// TODO: Authorize sub realms to be able to call protected functions (use DAOs)

// Keeps the package path to Boards2 realm.
// It ignores the last realm path element which is the current version.
var boardsNS = path.Dir(runtime.CurrentRealm().PkgPath()) + "/"

// GetRealmPermissions returns Boards2 realm permissions.
// This is a protected function only callable by Boards2 sub realms.
func GetRealmPermissions() boards.Permissions {
	assertRealmHasBoardsNS(runtime.CurrentRealm())
	return gPerms
}

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

// Iterate iterates boards.
// Iteration is done for all boards, including the ones that are not listed.
// To reverse iterate boards use a negative count.
// If the callback returns true, iteration is stopped.
func Iterate(start, count int, fn boards.BoardIterFn) bool {
	assertRealmHasBoardsNS(runtime.CurrentRealm())
	// TODO: We should check that current realm is authorized to iterate boards
	//       Realm must be reviewed before being authorized, to make sure boards
	//       are actually independent and won't be modified indirectly.
	return gBoards.Iterate(start, -count, fn)
}

// assertRealmHasBoardsNS asserts that a realm lives within Boards2 namespace.
func assertRealmHasBoardsNS(r runtime.Realm) {
	if !strings.HasPrefix(r.PkgPath(), boardsNS) {
		panic("forbidden")
	}
}
