package boards2

import (
	"chain"
	"chain/runtime"
	"strconv"

	"gno.land/p/gnoland/boards"
)

// FreezeBoard freezes a board so no more threads and comments can be created or modified.
func FreezeBoard(_ realm, boardID boards.ID) {
	setBoardReadonly(boardID, true)
}

// UnfreezeBoard removes frozen status from a board.
func UnfreezeBoard(_ realm, boardID boards.ID) {
	setBoardReadonly(boardID, false)
}

// IsBoardFrozen checks if a board has been frozen.
func IsBoardFrozen(boardID boards.ID) bool {
	board := mustGetBoard(boardID)
	return board.Readonly
}

// FreezeThread freezes a thread so thread cannot be replied, modified or deleted.
//
// Fails if board is frozen.
func FreezeThread(_ realm, boardID, threadID boards.ID) {
	setThreadReadonly(boardID, threadID, true)
}

// UnfreezeThread removes frozen status from a thread.
//
// Fails if board is frozen.
func UnfreezeThread(_ realm, boardID, threadID boards.ID) {
	setThreadReadonly(boardID, threadID, false)
}

// IsThreadFrozen checks if a thread has been frozen.
//
// Returns true if board is frozen.
func IsThreadFrozen(boardID, threadID boards.ID) bool {
	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	return board.Readonly || thread.Readonly
}

func setBoardReadonly(boardID boards.ID, readonly bool) {
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	if readonly {
		assertBoardIsNotFrozen(board)
	}

	caller := runtime.PreviousRealm().Address()
	args := boards.Args{caller, board.ID, readonly}
	board.Permissions.WithPermission(cross, caller, PermissionBoardFreeze, args, func(realm) {
		assertRealmIsNotLocked()

		board.Readonly = readonly

		chain.Emit(
			"BoardFreeze",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"frozen", strconv.FormatBool(readonly),
		)
	})
}

func setThreadReadonly(boardID, threadID boards.ID, readonly bool) {
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	if readonly {
		assertThreadIsNotFrozen(thread)
	}

	caller := runtime.PreviousRealm().Address()
	args := boards.Args{caller, board.ID, thread.ID, readonly}
	board.Permissions.WithPermission(cross, caller, PermissionThreadFreeze, args, func(realm) {
		assertRealmIsNotLocked()

		thread.Readonly = readonly

		chain.Emit(
			"ThreadFreeze",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
			"frozen", strconv.FormatBool(readonly),
		)
	})
}
