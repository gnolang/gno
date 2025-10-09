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
	assertThreadIsVisible(thread)

	return board.Readonly || thread.Readonly
}

// UnfreezeReply removes frozen status from a reply.
//
// Fails when parent thread or board are frozen.
func UnfreezeReply(_ realm, boardID, threadID, replyID boards.ID) {
	// TODO: Remove reply freezing
	setReplyReadonly(boardID, threadID, replyID, false)
}

// FreezeReply freezes a thread reply so it cannot be modified or deleted.
//
// Fails when parent thread or board are frozen.
func FreezeReply(_ realm, boardID, threadID, replyID boards.ID) {
	setReplyReadonly(boardID, threadID, replyID, true)
}

// IsReplyFrozen checks if a thread reply has been frozen.
//
// Returns true when board or a parent thread is frozen.
func IsReplyFrozen(boardID, threadID, replyID boards.ID) bool {
	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)

	return board.Readonly || thread.Readonly || reply.Readonly
}

func setReplyReadonly(boardID, threadID, replyID boards.ID, readonly bool) {
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := runtime.PreviousRealm().Address()
	assertHasPermission(board.Permissions, caller, PermissionReplyFreeze)

	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)

	if readonly {
		assertReplyIsNotFrozen(reply)
	}

	reply.Readonly = readonly

	chain.Emit(
		"ReplyFreeze",
		"caller", caller.String(),
		"boardID", board.ID.String(),
		"threadID", thread.ID.String(),
		"replyID", reply.ID.String(),
		"frozen", strconv.FormatBool(readonly),
	)
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
