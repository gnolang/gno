package boards2

import (
	"chain"
	"strconv"

	"gno.land/p/gnoland/boards"
)

// FreezeBoard freezes a board so no more threads and comments can be created or modified.
func FreezeBoard(cur realm, boardID boards.ID) {
	setBoardReadonly(0, cur, boardID, true)
}

// UnfreezeBoard removes frozen status from a board.
func UnfreezeBoard(cur realm, boardID boards.ID) {
	setBoardReadonly(0, cur, boardID, false)
}

// IsBoardFrozen checks if a board has been frozen.
func IsBoardFrozen(boardID boards.ID) bool {
	board := mustGetBoard(boardID)
	return board.Readonly
}

// FreezeThread freezes a thread so thread cannot be replied, modified or deleted.
//
// Fails if board is frozen.
func FreezeThread(cur realm, boardID, threadID boards.ID) {
	setThreadReadonly(0, cur, boardID, threadID, true)
}

// UnfreezeThread removes frozen status from a thread.
//
// Fails if board is frozen.
func UnfreezeThread(cur realm, boardID, threadID boards.ID) {
	setThreadReadonly(0, cur, boardID, threadID, false)
}

// IsThreadFrozen checks if a thread has been frozen.
//
// Returns true if board is frozen.
func IsThreadFrozen(boardID, threadID boards.ID) bool {
	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	return board.Readonly || thread.Readonly
}

func setBoardReadonly(_ int, rlm realm, boardID boards.ID, readonly bool) {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	if readonly {
		assertBoardIsNotFrozen(board)
	}

	caller := rlm.Previous().Address()
	args := boards.Args{caller, board.ID, readonly}
	board.Permissions.WithPermission(caller, PermissionBoardFreeze, args, func() {
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

func setThreadReadonly(_ int, rlm realm, boardID, threadID boards.ID, readonly bool) {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	if readonly {
		assertThreadIsNotFrozen(thread)
	}

	caller := rlm.Previous().Address()
	args := boards.Args{caller, board.ID, thread.ID, readonly}
	board.Permissions.WithPermission(caller, PermissionThreadFreeze, args, func() {
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
