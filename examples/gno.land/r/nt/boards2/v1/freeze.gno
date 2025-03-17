package boards2

import (
	"std"
	"strconv"
)

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
