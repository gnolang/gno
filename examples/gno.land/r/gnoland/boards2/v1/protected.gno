// The `protected.gno` file contains public realm functions that expose Boards2
// data to other realms. This allows sub realms to access boards data to be able
// to migrate content from one version to another or to implement specific
// features in separate sub realms.

package boards2

import (
	"gno.land/p/gnoland/boards"
)

// TODO: Authorize sub realms to be able to call protected functions (use DAOs)

// GetRealmPermissions returns Boards2 realm permissions.
func GetRealmPermissions() boards.Permissions {
	return gPerms
}

// GetBoard returns a board.
func GetBoard(boardID boards.ID) (_ *boards.Board, found bool) {
	return gBoards.Get(boardID)
}

// MustGetBoard returns a board or panics on error.
func MustGetBoard(boardID boards.ID) *boards.Board {
	return mustGetBoard(boardID)
}

// Iterate iterates boards.
// Iteration is done for all boards, including the ones that are not listed.
// To reverse iterate boards use a negative count.
// If the callback returns true, iteration is stopped.
func Iterate(start, count int, fn boards.BoardIterFn) bool {
	return gBoards.Iterate(start, count, fn)
}
