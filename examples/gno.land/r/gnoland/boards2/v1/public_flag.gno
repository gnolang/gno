package boards2

import (
	"chain"
	"chain/runtime"
	"strconv"

	"gno.land/p/gnoland/boards"
)

// SetFlaggingThreshold sets the number of flags required to hide a thread or comment.
//
// Threshold is only applicable within the board where it's setted.
func SetFlaggingThreshold(_ realm, boardID boards.ID, threshold int) {
	if threshold < 1 {
		panic("invalid flagging threshold")
	}

	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := runtime.PreviousRealm().Address()
	args := boards.Args{boardID, threshold}
	board.Permissions.WithPermission(cross, caller, PermissionBoardFlaggingUpdate, args, func(realm) {
		assertRealmIsNotLocked()

		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)

		gFlaggingThresholds.Set(boardID.String(), threshold)
		chain.Emit(
			"FlaggingThresholdUpdated",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threshold", strconv.Itoa(threshold),
		)
	})
}

// GetFlaggingThreshold returns the number of flags required to hide a thread or comment within a board.
func GetFlaggingThreshold(boardID boards.ID) int {
	assertBoardExists(boardID)
	return getFlaggingThreshold(boardID)
}

// FlagThread adds a new flag to a thread.
//
// Flagging requires special permissions and hides the thread when
// the number of flags reaches a pre-defined flagging threshold.
func FlagThread(_ realm, boardID, threadID boards.ID, reason string) {
	board := mustGetBoard(boardID)

	// Realm owners should be able to flag without permissions even when board is frozen
	caller := runtime.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
		assertHasPermission(board.Permissions, caller, PermissionThreadFlag)
	}

	thread, found := board.Threads.Get(threadID)
	if !found {
		panic("thread not found")
	}

	assertThreadIsNotFrozen(thread)

	// Realm owners can hide with a single flag
	hide := flagItem(thread, caller, reason, getFlaggingThreshold(boardID))
	if hide || isRealmOwner {
		thread.Hidden = true
	}

	chain.Emit(
		"ThreadFlagged",
		"caller", caller.String(),
		"boardID", board.ID.String(),
		"threadID", thread.ID.String(),
		"reason", reason,
	)
}

// FlagReply adds a new flag to a comment or reply.
//
// Flagging requires special permissions and hides the comment or reply
// when the number of flags reaches a pre-defined flagging threshold.
func FlagReply(_ realm, boardID, threadID, replyID boards.ID, reason string) {
	board := mustGetBoard(boardID)

	// Realm owners should be able to flag without permissions even when board is frozen
	caller := runtime.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
		assertHasPermission(board.Permissions, caller, PermissionReplyFlag)
	}

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsNotFrozen(thread)

	// Realm owners can hide with a single flag
	hide := flagItem(reply, caller, reason, getFlaggingThreshold(boardID))
	if hide || isRealmOwner {
		reply.Hidden = true
	}

	chain.Emit(
		"ReplyFlagged",
		"caller", caller.String(),
		"boardID", board.ID.String(),
		"threadID", thread.ID.String(),
		"replyID", reply.ID.String(),
		"reason", reason,
	)
}
