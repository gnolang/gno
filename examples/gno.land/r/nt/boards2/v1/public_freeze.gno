package boards2

import (
	"std"
	"strconv"
)

// FreezeBoard freezes a board so no more threads and comments can be created or modified.
func FreezeBoard(boardID BoardID) {
	setBoardReadOnly(boardID, true)
}

// UnfreezeBoard removes frozen status from a board.
func UnfreezeBoard(boardID BoardID) {
	setBoardReadOnly(boardID, false)
}

// IsBoardFrozen checks if a board has been frozen.
func IsBoardFrozen(boardID BoardID) bool {
	board := mustGetBoard(boardID)
	return board.IsReadOnly()
}

// FreezeThread freezes a thread so thread cannot be replied, modified or deleted.
//
// Fails if board is frozen.
func FreezeThread(boardID BoardID, threadID PostID) {
	setThreadReadOnly(boardID, threadID, true)
}

// UnfreezeThread removes frozen status from a thread.
//
// Fails if board is frozen.
func UnfreezeThread(boardID BoardID, threadID PostID) {
	setThreadReadOnly(boardID, threadID, false)
}

// IsThreadFrozen checks if a thread has been frozen.
//
// Returns true if board is frozen.
func IsThreadFrozen(boardID BoardID, threadID PostID) bool {
	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	return board.IsReadOnly() || thread.IsReadOnly()
}

// UnfreezeReply removes frozen status from a reply.
//
// Fails when parent thread or board are frozen.
func UnfreezeReply(boardID BoardID, threadID, replyID PostID) {
	// XXX: Is there a use case for also freezing replies?
	setReplyReadOnly(boardID, threadID, replyID, false)
}

// FreezeReply freezes a thread reply so it cannot be modified or deleted.
//
// Fails when parent thread or board are frozen.
func FreezeReply(boardID BoardID, threadID, replyID PostID) {
	setReplyReadOnly(boardID, threadID, replyID, true)
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

	return board.IsReadOnly() || thread.IsReadOnly() || reply.IsReadOnly()
}

func setReplyReadOnly(boardID BoardID, threadID, replyID PostID, isReadOnly bool) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionReplyFreeze)

	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)

	if isReadOnly {
		assertReplyIsNotFrozen(reply)
	}

	reply.SetReadOnly(isReadOnly)

	std.Emit(
		"ReplyFreeze",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", replyID.String(),
		"frozen", strconv.FormatBool(isReadOnly),
	)
}

func setThreadReadOnly(boardID BoardID, threadID PostID, isReadOnly bool) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionThreadFreeze)

	thread := mustGetThread(board, threadID)
	if isReadOnly {
		assertThreadIsNotFrozen(thread)
	}

	thread.SetReadOnly(isReadOnly)

	std.Emit(
		"ThreadFreeze",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"frozen", strconv.FormatBool(isReadOnly),
	)
}

func setBoardReadOnly(boardID BoardID, isReadOnly bool) {
	board := mustGetBoard(boardID)
	if isReadOnly {
		assertBoardIsNotFrozen(board)
	}

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionBoardFreeze)

	board.SetReadOnly(isReadOnly)

	std.Emit(
		"BoardFreeze",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"frozen", strconv.FormatBool(isReadOnly),
	)
}
