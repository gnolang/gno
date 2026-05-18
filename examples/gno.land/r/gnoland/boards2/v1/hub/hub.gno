// Package hub exposes safe, read-only views over boards2's persistent
// state. The Get* functions are crossing functions: callers must invoke
// them with `cross` (e.g. `hub.GetBoard(cross, id)`). The crossing into
// hub gives each call a live `cur` whose `Previous()` is the caller's
// realm — that's what boards2's protected entry points use for the
// namespace check.
//
// Note for future maintainers: inside a Get* body, `cur.Previous()` is
// the *immediate* caller's realm, which may be an intermediary, not
// necessarily the end user. Do not graft user-identity gating onto
// these reads.
package hub

import (
	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

// GetBoard returns a safe board.
func GetBoard(cur realm, id uint64) (Board, bool) {
	b, found := getBoard(0, cur, id)
	if !found {
		return Board{}, false
	}
	return NewSafeBoard(b), true
}

// GetThread returns a safe board thread.
func GetThread(cur realm, boardID, threadID uint64) (Thread, bool) {
	t, found := getThread(0, cur, boardID, threadID)
	if !found {
		return Thread{}, false
	}
	return NewSafeThread(t), true
}

// GetComment returns a safe thread comment.
func GetComment(cur realm, boardID, threadID, commentID uint64) (Comment, bool) {
	c, found := getComment(0, cur, boardID, threadID, commentID)
	if !found {
		return Comment{}, false
	}
	return NewSafeComment(c), true
}

// GetBoards returns a list with all boards.
// To reverse iterate use a negative count.
func GetBoards(cur realm, start, count int) []Board {
	var boards_ []Board
	boards2.Iterate(cross2(cur), start, count, func(b *boards.Board) bool {
		boards_ = append(boards_, NewSafeBoard(b))
		return false
	})
	return boards_
}

// GetThreads returns a list with threads of a board.
// To reverse iterate use a negative count.
func GetThreads(cur realm, boardID uint64, start, count int) []Thread {
	b, found := getBoard(0, cur, boardID)
	if !found {
		return nil
	}

	var threads []Thread
	b.Threads.Iterate(start, count, func(thread *boards.Post) bool {
		threads = append(threads, NewSafeThread(thread))
		return false
	})
	return threads
}

// GetMembers returns a list with the members of a board.
// To reverse iterate use a negative count.
func GetMembers(cur realm, boardID uint64, start, count int) []boards.User {
	b, found := getBoard(0, cur, boardID)
	if !found {
		return nil
	}

	var members []boards.User
	b.Permissions.IterateUsers(start, count, func(u boards.User) bool {
		members = append(members, u)
		return false
	})
	return members
}

// GetReposts returns a list with repost of a board thread.
// To reverse iterate use a negative count.
func GetReposts(cur realm, boardID, threadID uint64, start, count int) []Thread {
	t, found := getThread(0, cur, boardID, threadID)
	if !found {
		return nil
	}

	var reposts []Thread
	t.Reposts.Iterate(start, count, func(rBoardID, rRepostID boards.ID) bool {
		r, found := getThread(0, cur, uint64(rBoardID), uint64(rRepostID))
		if found {
			reposts = append(reposts, NewSafeThread(r))
		}
		return false
	})
	return reposts
}

// GetFlag returns a list with thread or comment moderation flags.
// To reverse iterate use a negative count.
// Thread flags are returned when `commentID` is zero, or comment flags are returned otherwise.
func GetFlags(cur realm, boardID, threadID, commentID uint64, start, count int) []boards.Flag {
	var storage boards.FlagStorage
	if commentID == 0 {
		t, found := getThread(0, cur, boardID, threadID)
		if !found {
			return nil
		}

		storage = t.Flags
	} else {
		c, found := getComment(0, cur, boardID, threadID, commentID)
		if !found {
			return nil
		}

		storage = c.Flags
	}

	var flags []boards.Flag
	storage.Iterate(start, count, func(f boards.Flag) bool {
		flags = append(flags, f)
		return false
	})
	return flags
}

// GetComments returns a list with all thread comments and replies.
// To reverse iterate use a negative count.
// Top level comments can be filtered by checking `Comment.ParentID`, replies
// always have a parent comment or reply, while comments have no parent.
func GetComments(cur realm, boardID, threadID uint64, start, count int) []Comment {
	t, found := getThread(0, cur, boardID, threadID)
	if !found {
		return nil
	}

	var comments []Comment
	t.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		comments = append(comments, NewSafeComment(comment))
		return false
	})
	return comments
}

// GetReplies returns a list with top level comment replies.
// To reverse iterate use a negative count.
func GetReplies(cur realm, boardID, threadID, commentID uint64, start, count int) []Comment {
	c, found := getComment(0, cur, boardID, threadID, commentID)
	if !found {
		return nil
	}

	var replies []Comment
	c.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		replies = append(replies, NewSafeComment(comment))
		return false
	})
	return replies
}

// getBoard fetches a raw *boards.Board via boards2's protected
// GetBoard. Non-crossing helper: rlm is hub's own cur, threaded
// through from the enclosing crossing function.
func getBoard(_ int, rlm realm, boardID uint64) (*boards.Board, bool) {
	return boards2.GetBoard(cross2(rlm), boards.ID(boardID))
}

func getThread(_ int, rlm realm, boardID, threadID uint64) (*boards.Post, bool) {
	b, found := getBoard(0, rlm, boardID)
	if !found {
		return nil, false
	}

	t, found := b.Threads.Get(boards.ID(threadID))
	if !found {
		// When thread is not found search it within hidden threads
		meta := b.Meta.(*boards2.BoardMeta)
		t, found = meta.HiddenThreads.Get(boards.ID(threadID))
	}
	return t, found
}

func getComment(_ int, rlm realm, boardID, threadID, commentID uint64) (*boards.Post, bool) {
	t, found := getThread(0, rlm, boardID, threadID)
	if !found {
		return nil, false
	}

	return t.Replies.Get(boards.ID(commentID))
}
