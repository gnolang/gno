// The `protected.gno` file contains public realm functions that must only be called
// from realms that live within the Boards2 package namespace. This allows sub realms
// to access boards data to be able to migrate content from one version to another or
// to implement specific features in separate sub realms.
package boards2

import (
	"strings"

	"gno.land/p/gnoland/boards"
)

// TODO: Authorize sub realms to be able to call protected functions (use DAOs)

// boardsNS is the package-path namespace shared by all Boards2 versions.
// Update as part of each version bump.
const boardsNS = "gno.land/r/gnoland/boards2/"

// GetRealmPermissions returns Boards2 realm permissions.
// This is a protected function only callable by Boards2 sub realms.
func GetRealmPermissions(cur realm) boards.Permissions {
	assertCallerHasBoardsNS(0, cur.Previous())
	return gPerms
}

// GetBoard returns a board.
// This is a protected function only callable by Boards2 sub realms.
func GetBoard(cur realm, boardID boards.ID) (_ *boards.Board, found bool) {
	assertCallerHasBoardsNS(0, cur.Previous())
	return gBoards.Get(boardID)
}

// MustGetBoard returns a board or panics on error.
// This is a protected function only callable by Boards2 sub realms.
func MustGetBoard(cur realm, boardID boards.ID) *boards.Board {
	assertCallerHasBoardsNS(0, cur.Previous())
	return mustGetBoard(boardID)
}

// Iterate iterates boards.
// Iteration is done for all boards, including the ones that are not listed.
// To reverse iterate boards use a negative count.
// If the callback returns true, iteration is stopped.
func Iterate(cur realm, start, count int, fn boards.BoardIterFn) bool {
	assertCallerHasBoardsNS(0, cur.Previous())
	return gBoards.Iterate(start, count, fn)
}

// assertCallerHasBoardsNS panics unless rlm.PkgPath() lives within the
// Boards2 namespace. Callers pass cur.Previous() — cross2 already validated
// cur as the immediate caller's live cur, so PkgPath() here is authentic.
// _ int discriminator keeps this helper non-crossing.
func assertCallerHasBoardsNS(_ int, rlm realm) {
	if !strings.HasPrefix(rlm.PkgPath(), boardsNS) {
		panic("forbidden, caller should live within \"" + boardsNS + "\" namespace")
	}
}
