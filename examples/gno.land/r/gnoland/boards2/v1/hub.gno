// The `hub.gno` file exposes safe, read-only views over the realm's
// persistent state.
//
// Note for future maintainers: these reads perform no caller
// authorization, so do not graft user-identity gating onto them.

package boards2

import (
	"gno.land/p/gnoland/boards"
	hubexts "gno.land/p/gnoland/boards/exts/hub"
)

// Safe view types are defined in the hub extensions package and
// re-exported here so callers can keep using them via this realm.
type (
	Board   = hubexts.Board
	Comment = hubexts.Comment
	Flag    = hubexts.Flag
	Member  = hubexts.Member
	Thread  = hubexts.Thread
)

// GetBoard returns a safe board.
func GetBoard(id uint64) (Board, bool) {
	b, found := gBoards.Get(boards.ID(id))
	if !found {
		return Board{}, false
	}
	return hubexts.NewSafeBoard(b), true
}

// GetThread returns a safe board thread.
func GetThread(boardID, threadID uint64) (Thread, bool) {
	t, found := getBoardThread(boardID, threadID)
	if !found {
		return Thread{}, false
	}
	return hubexts.NewSafeThread(t), true
}

// GetComment returns a safe thread comment.
func GetComment(boardID, threadID, commentID uint64) (Comment, bool) {
	c, found := getComment(boardID, threadID, commentID)
	if !found {
		return Comment{}, false
	}
	return hubexts.NewSafeComment(c), true
}

// GetBoards returns a list with all boards.
// To reverse iterate use a negative count.
func GetBoards(start, count int) []Board {
	var boards_ []Board
	gBoards.Iterate(start, count, func(b *boards.Board) bool {
		boards_ = append(boards_, hubexts.NewSafeBoard(b))
		return false
	})
	return boards_
}

// GetThreads returns a list with threads of a board.
// To reverse iterate use a negative count.
func GetThreads(boardID uint64, start, count int) []Thread {
	b, found := gBoards.Get(boards.ID(boardID))
	if !found {
		return nil
	}

	var threads []Thread
	b.Threads.Iterate(start, count, func(thread *boards.Post) bool {
		threads = append(threads, hubexts.NewSafeThread(thread))
		return false
	})
	return threads
}

// GetMembers returns a list with the members of a board.
// To reverse iterate use a negative count.
func GetMembers(boardID uint64, start, count int) []boards.User {
	b, found := gBoards.Get(boards.ID(boardID))
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
func GetReposts(boardID, threadID uint64, start, count int) []Thread {
	t, found := getBoardThread(boardID, threadID)
	if !found {
		return nil
	}

	var reposts []Thread
	t.Reposts.Iterate(start, count, func(rBoardID, rRepostID boards.ID) bool {
		r, found := getBoardThread(uint64(rBoardID), uint64(rRepostID))
		if found {
			reposts = append(reposts, hubexts.NewSafeThread(r))
		}
		return false
	})
	return reposts
}

// GetFlag returns a list with thread or comment moderation flags.
// To reverse iterate use a negative count.
// Thread flags are returned when `commentID` is zero, or comment flags are returned otherwise.
func GetFlags(boardID, threadID, commentID uint64, start, count int) []boards.Flag {
	var storage boards.FlagStorage
	if commentID == 0 {
		t, found := getBoardThread(boardID, threadID)
		if !found {
			return nil
		}

		storage = t.Flags
	} else {
		c, found := getComment(boardID, threadID, commentID)
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
func GetComments(boardID, threadID uint64, start, count int) []Comment {
	t, found := getBoardThread(boardID, threadID)
	if !found {
		return nil
	}

	var comments []Comment
	t.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		comments = append(comments, hubexts.NewSafeComment(comment))
		return false
	})
	return comments
}

// GetReplies returns a list with top level comment replies.
// To reverse iterate use a negative count.
func GetReplies(boardID, threadID, commentID uint64, start, count int) []Comment {
	c, found := getComment(boardID, threadID, commentID)
	if !found {
		return nil
	}

	var replies []Comment
	c.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		replies = append(replies, hubexts.NewSafeComment(comment))
		return false
	})
	return replies
}

// getBoardThread returns a board thread from their IDs.
func getBoardThread(boardID, threadID uint64) (*boards.Post, bool) {
	b, found := gBoards.Get(boards.ID(boardID))
	if !found {
		return nil, false
	}
	return getThread(b, boards.ID(threadID))
}

// getComment returns a thread comment from their IDs.
func getComment(boardID, threadID, commentID uint64) (*boards.Post, bool) {
	t, found := getBoardThread(boardID, threadID)
	if !found {
		return nil, false
	}

	return t.Replies.Get(boards.ID(commentID))
}
