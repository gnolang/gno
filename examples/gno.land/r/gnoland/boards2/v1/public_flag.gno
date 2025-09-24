package boards2

import (
	"std"
	"strconv"
)

// SetFlaggingThreshold sets the number of flags required to hide a thread or comment.
//
// Threshold is only applicable within the board where it's setted.
func SetFlaggingThreshold(_ realm, boardID BoardID, threshold int) {
	if threshold < 1 {
		panic("invalid flagging threshold")
	}

	assertRealmIsNotLocked()

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	args := Args{boardID, threshold}
	board.perms.WithPermission(cross, caller, PermissionBoardFlaggingUpdate, args, func(realm, Args) {
		assertRealmIsNotLocked()

		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)

		gFlaggingThresholds.Set(boardID.String(), threshold)
		std.Emit(
			"FlaggingThresholdUpdated",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"threshold", strconv.Itoa(threshold),
		)
	})
}

// GetFlaggingThreshold returns the number of flags required to hide a thread or comment within a board.
func GetFlaggingThreshold(boardID BoardID) int {
	assertBoardExists(boardID)
	return getFlaggingThreshold(boardID)
}

// FlagThread adds a new flag to a thread.
//
// Flagging requires special permissions and hides the thread when
// the number of flags reaches a pre-defined flagging threshold.
func FlagThread(_ realm, boardID BoardID, threadID PostID, reason string) {
	board := mustGetBoard(boardID)

	// Realm owners should be able to flag without permissions even when board is frozen
	caller := std.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
		assertHasPermission(board.perms, caller, PermissionThreadFlag)
	}

	t, ok := board.GetThread(threadID)
	if !ok {
		panic("post doesn't exist")
	}
	assertThreadIsNotFrozen(t)

	// Realm owners can hide with a single flag
	hide := flagItem(t, caller, reason, getFlaggingThreshold(boardID))
	if hide || isRealmOwner {
		t.Hidden = true
	}

	std.Emit(
		"ThreadFlagged",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"reason", reason,
	)
}

// FlagReply adds a new flag to a comment or reply.
//
// Flagging requires special permissions and hides the comment or reply
// when the number of flags reaches a pre-defined flagging threshold.
func FlagReply(_ realm, boardID BoardID, threadID, replyID PostID, reason string) {
	board := mustGetBoard(boardID)

	// Realm owners should be able to flag without permissions even when board is frozen
	caller := std.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
		assertHasPermission(board.perms, caller, PermissionReplyFlag)
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

	std.Emit(
		"ReplyFlagged",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", replyID.String(),
		"reason", reason,
	)
}
