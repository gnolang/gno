package boards2

import (
	"std"
	"strconv"
)

type freezeArgs struct {
	boardID  BoardID
	threadID PostID
	replyID  PostID
}

func (args freezeArgs) getThread() (*Post, *Board) {
	board := mustGetBoard(args.boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, args.threadID)
	return thread, board
}

func (args freezeArgs) getReply() (*Post, *Board) {
	board := mustGetBoard(args.boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, args.threadID)
	assertThreadVisible(thread)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, args.replyID)
	assertReplyVisible(reply)
	return reply, board
}

func setReplyReadOnly(args freezeArgs, isReadOnly bool) {
	reply, board := args.getReply()

	if isReadOnly {
		// disallow freezing already frozen reply
		assertReplyIsNotFrozen(reply)
	}

	caller := std.OriginCaller()
	permArgs := Args{args.boardID, args.threadID, args.replyID}
	board.perms.WithPermission(caller, PermissionReplyFreeze, permArgs, func(Args) {
		reply, _ := args.getReply()
		reply.SetReadOnly(isReadOnly)

		std.Emit(
			"ReplyFreeze",
			"caller", caller.String(),
			"boardID", args.boardID.String(),
			"threadID", args.threadID.String(),
			"replyID", args.replyID.String(),
			"frozen", strconv.FormatBool(isReadOnly),
		)
	})
}

func setThreadReadOnly(args freezeArgs, isReadOnly bool) {
	thread, board := args.getThread()

	if isReadOnly {
		// disallow freezing of already frozen thread
		assertThreadIsNotFrozen(thread)
	}

	caller := std.OriginCaller()
	permArgs := Args{args.boardID, args.threadID}
	board.perms.WithPermission(caller, PermissionThreadFreeze, permArgs, func(Args) {
		thread, _ := args.getThread()
		thread.SetReadOnly(isReadOnly)

		std.Emit(
			"ThreadFreeze",
			"caller", caller.String(),
			"boardID", args.boardID.String(),
			"threadID", args.threadID.String(),
			"frozen", strconv.FormatBool(isReadOnly),
		)
	})
}

func setBoardReadOnly(boardID BoardID, isReadOnly bool) {
	board := mustGetBoard(boardID)
	if isReadOnly {
		assertBoardIsNotFrozen(board)
	}

	caller := std.OriginCaller()
	args := Args{boardID}
	board.perms.WithPermission(caller, PermissionBoardFreeze, args, func(Args) {
		board := mustGetBoard(boardID)
		board.SetReadOnly(isReadOnly)

		std.Emit(
			"BoardFreeze",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"frozen", strconv.FormatBool(isReadOnly),
		)
	})
}
