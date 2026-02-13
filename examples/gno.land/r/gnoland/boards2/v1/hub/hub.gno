package hub

import (
	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

// GetBoard returns a safe board.
func GetBoard(id uint64) (Board, bool) {
	b, found := boards2.GetBoard(boards.ID(id))
	if !found {
		return Board{}, false
	}
	return newSafeBoard(b), true
}

// GetThread returns a safe board thread.
func GetThread(boardID, threadID uint64) (Thread, bool) {
	t, found := getThread(boardID, threadID)
	if !found {
		return Thread{}, false
	}
	return newSafeThread(t), true
}

// GetBoards returns a list with all boards.
// To reverse iterate use a negative count.
func GetBoards(start, count int) []Board {
	var boards_ []Board
	boards2.Iterate(start, count, func(b *boards.Board) bool {
		boards_ = append(boards_, newSafeBoard(b))
		return false
	})
	return boards_
}

// GetThreads returns a list with threads of a board.
// To reverse iterate use a negative count.
func GetThreads(boardID uint64, start, count int) []Thread {
	b, found := boards2.GetBoard(boards.ID(boardID))
	if !found {
		return nil
	}

	var threads []Thread
	b.Threads.Iterate(start, count, func(thread *boards.Post) bool {
		threads = append(threads, newSafeThread(thread))
		return false
	})
	return threads
}

// GetMembers returns a list with the members of a board.
// To reverse iterate use a negative count.
func GetMembers(boardID uint64, start, count int) []boards.User {
	b, found := boards2.GetBoard(boards.ID(boardID))
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
	t, found := getThread(boardID, threadID)
	if !found {
		return nil
	}

	var reposts []Thread
	t.Reposts.Iterate(start, count, func(boardID, repostID boards.ID) bool {
		r, found := GetThread(uint64(boardID), uint64(repostID))
		if found {
			reposts = append(reposts, r)
		}
		return false
	})
	return reposts
}

// GetFlag returns a list with thread moderation flags.
// To reverse iterate use a negative count.
func GetFlags(boardID, threadID uint64, start, count int) []boards.Flag {
	t, found := getThread(boardID, threadID)
	if !found {
		return nil
	}

	var flags []boards.Flag
	t.Flags.Iterate(start, count, func(f boards.Flag) bool {
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
	t, found := getThread(boardID, threadID)
	if !found {
		return nil
	}

	var comments []Comment
	t.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		comments = append(comments, newSafeComment(comment))
		return false
	})
	return comments
}

// GetReplies returns a list with top level comment replies.
// To reverse iterate use a negative count.
// Use `GetComments()` to get all thread comments and replies.
func GetReplies(boardID, threadID, commentID uint64, start, count int) []Comment {
	c, found := getComment(boardID, threadID, commentID)
	if !found {
		return nil
	}

	var replies []Comment
	c.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		replies = append(replies, newSafeComment(comment))
		return false
	})
	return replies
}

func getThread(boardID, threadID uint64) (*boards.Post, bool) {
	b, found := boards2.GetBoard(boards.ID(boardID))
	if !found {
		return nil, false
	}

	return b.Threads.Get(boards.ID(threadID))
}

func getComment(boardID, threadID, commentID uint64) (*boards.Post, bool) {
	t, found := getThread(boardID, threadID)
	if !found {
		return nil, false
	}

	return t.Replies.Get(boards.ID(commentID))
}
