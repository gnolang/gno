package boards2

import (
	"chain"
	"chain/runtime"
	"strconv"
	"strings"

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
	args := boards.Args{board.ID, threshold}
	board.Permissions.WithPermission(caller, PermissionBoardFlaggingUpdate, args, crossingFn(func() {
		assertRealmIsNotLocked()

		gFlaggingThresholds.Set(boardID.String(), threshold)

		chain.Emit(
			"FlaggingThresholdUpdated",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threshold", strconv.Itoa(threshold),
		)
	}))
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
	reason = strings.TrimSpace(reason)
	if reason == "" {
		panic("flagging reason is required")
	}

	caller := runtime.PreviousRealm().Address()
	board := mustGetBoard(boardID)
	isRealmOwner := gPerms.HasRole(caller, RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
	}

	thread, found := getThread(board, threadID)
	if !found {
		panic("thread not found")
	}

	if thread.Hidden {
		panic("flagging hidden threads is not allowed")
	}

	flagThread := func() {
		if thread.Hidden {
			panic("flagged thread is already hidden")
		}

		// Hide thread when flagging threshold is reached.
		// Realm owners can hide with a single flag.
		hide := flagItem(thread, caller, reason, getFlaggingThreshold(board.ID))
		if hide || isRealmOwner {
			// Remove thread from the list of visible threads
			thread, removed := board.Threads.Remove(threadID)
			if !removed {
				panic("thread not found")
			}

			// Mark thread as hidden to avoid rendering content
			thread.Hidden = true

			// Keep track of hidden the thread to be able to restore it after moderation disputes
			meta := board.Meta.(*Meta)
			meta.HiddenThreads.Add(thread)
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
	board.Permissions.WithPermission(caller, PermissionThreadFlag, args, crossingFn(func() {
		flagThread()
	}))
}

// FlagReply adds a new flag to a comment or reply.
//
// Flagging requires special permissions and hides the comment or reply
// when the number of flags reaches a pre-defined flagging threshold.
func FlagReply(_ realm, boardID, threadID, replyID boards.ID, reason string) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		panic("flagging reason is required")
	}

	caller := runtime.PreviousRealm().Address()
	board := mustGetBoard(boardID)
	isRealmOwner := gPerms.HasRole(caller, RoleOwner)
	if !isRealmOwner {
		assertRealmIsNotLocked()
		assertBoardIsNotFrozen(board)
	}

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	if reply.Hidden {
		panic("flagging hidden comments or replies is not allowed")
	}

	flagReply := func() {
		if reply.Hidden {
			panic("flagged comment or reply is already hidden")
		}

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
	board.Permissions.WithPermission(caller, PermissionReplyFlag, args, crossingFn(func() {
		flagReply()
	}))
}
