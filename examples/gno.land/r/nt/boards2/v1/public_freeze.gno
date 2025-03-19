package boards2

import (
	"std"
	"strconv"
)

// FreezeBoard freezes a board so no more threads and comments can be created or modified.
func FreezeBoard(boardID BoardID) {
	setBoardReadonly(boardID, true)
}

// UnfreezeBoard removes frozen status from a board.
func UnfreezeBoard(boardID BoardID) {
	setBoardReadonly(boardID, false)
}

// IsBoardFrozen checks if a board has been frozen.
func IsBoardFrozen(boardID BoardID) bool {
	board := mustGetBoard(boardID)
	return board.IsReadonly()
}

// FreezeThread freezes a thread so thread cannot be replied, modified or deleted.
//
// Fails if board is frozen.
func FreezeThread(boardID BoardID, threadID PostID) {
	setThreadReadonly(boardID, threadID, true)
}

// UnfreezeThread removes frozen status from a thread.
//
// Fails if board is frozen.
func UnfreezeThread(boardID BoardID, threadID PostID) {
	setThreadReadonly(boardID, threadID, false)
}

// IsThreadFrozen checks if a thread has been frozen.
//
// Returns true if board is frozen.
func IsThreadFrozen(boardID BoardID, threadID PostID) bool {
	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	return board.IsReadonly() || thread.IsReadonly()
}

// UnfreezeReply removes frozen status from a reply.
//
// Fails when parent thread or board are frozen.
func UnfreezeReply(boardID BoardID, threadID, replyID PostID) {
	// XXX: Is there a use case for also freezing replies?
	setReplyReadonly(boardID, threadID, replyID, false)
}

// FreezeReply freezes a thread reply so it cannot be modified or deleted.
//
// Fails when parent thread or board are frozen.
func FreezeReply(boardID BoardID, threadID, replyID PostID) {
	setReplyReadonly(boardID, threadID, replyID, true)
}

// IsReplyFrozen checks if a thread reply has been frozen.
//
// Returns true when board or a parent thread is frozen.
func IsReplyFrozen(boardID BoardID, threadID, replyID PostID) bool {
	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)

	return board.IsReadonly() || thread.IsReadonly() || reply.IsReadonly()
}

func setReplyReadonly(boardID BoardID, threadID, replyID PostID, isReadonly bool) {
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionReplyFreeze)

	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)

	if isReadonly {
		assertReplyIsNotFrozen(reply)
	}

	reply.SetReadonly(isReadonly)

	std.Emit(
		"ReplyFreeze",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", replyID.String(),
		"frozen", strconv.FormatBool(isReadonly),
	)
}

func setThreadReadonly(boardID BoardID, threadID PostID, isReadonly bool) {
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionThreadFreeze)

	thread := mustGetThread(board, threadID)
	if isReadonly {
		assertThreadIsNotFrozen(thread)
	}

	thread.SetReadonly(isReadonly)

	std.Emit(
		"ThreadFreeze",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"frozen", strconv.FormatBool(isReadonly),
	)
}

func setBoardReadonly(boardID BoardID, isReadonly bool) {
	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	if isReadonly {
		assertBoardIsNotFrozen(board)
	}

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionBoardFreeze)

	board.SetReadonly(isReadonly)

	std.Emit(
		"BoardFreeze",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"frozen", strconv.FormatBool(isReadonly),
	)
}
