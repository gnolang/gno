package boards2

import (
	"chain"
	"chain/runtime"
	"strconv"

	"gno.land/p/gnoland/boards"

	"gno.land/r/gnoland/boards2/v1/permissions"
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
	args := boards.Args{board.ID, threshold}
	board.Permissions.WithPermission(cross, caller, permissions.PermissionBoardFlaggingUpdate, args, func(realm) {
		assertRealmIsNotLocked()

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
	caller := runtime.PreviousRealm().Address()
	board := mustGetBoard(boardID)
	isRealmOwner := gPerms.HasRole(caller, permissions.RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
	}

	thread, found := board.Threads.Get(threadID)
	if !found {
		panic("thread not found")
	}

	assertThreadIsNotFrozen(thread)

	flagThread := func() {
		hide := flagItem(thread, caller, reason, getFlaggingThreshold(board.ID))
		if hide || isRealmOwner {
			// Realm owners can hide with a single flag
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

	// Realm owners should be able to flag without permissions even when board is frozen
	if isRealmOwner {
		flagThread()
		return
	}

	args := boards.Args{caller, board.ID, thread.ID, reason}
	board.Permissions.WithPermission(cross, caller, permissions.PermissionThreadFlag, args, func(realm) {
		flagThread()
	})
}

// FlagReply adds a new flag to a comment or reply.
//
// Flagging requires special permissions and hides the comment or reply
// when the number of flags reaches a pre-defined flagging threshold.
func FlagReply(_ realm, boardID, threadID, replyID boards.ID, reason string) {
	caller := runtime.PreviousRealm().Address()
	board := mustGetBoard(boardID)
	isRealmOwner := gPerms.HasRole(caller, permissions.RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
	}

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	flagReply := func() {
		hide := flagItem(reply, caller, reason, getFlaggingThreshold(board.ID))
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

	// Realm owners should be able to flag without permissions even when board is frozen
	if isRealmOwner {
		flagReply()
		return
	}

	args := boards.Args{caller, board.ID, thread.ID, reply.ID, reason}
	board.Permissions.WithPermission(cross, caller, permissions.PermissionReplyFlag, args, func(realm) {
		flagReply()
	})
}
