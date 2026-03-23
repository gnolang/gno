package hub

import (
	"gno.land/p/gnoland/boards"
)

// Flag defines a type for thread and comment flags.
type Flag struct {
	// User is the user that flagged.
	User address

	// Reason is the reason for flagging.
	Reason string
}

// Thread defines a type for board threads.
type Thread struct {
	ref *boards.Post

	// ID is the unique identifier of the thread.
	ID uint64

	// OriginalBoardID contains the board ID of the original thread when current is a repost.
	OriginalBoardID uint64

	// BoardID is the board ID where thread is created.
	BoardID uint64

	// Title contains thread's title.
	Title string

	// Body contains content of the thread.
	Body string

	// Hidden indicates that thread is hidden.
	Hidden bool

	// Readonly indicates that thread is readonly.
	Readonly bool

	// CommentCount contains the number of thread comments.
	// Count only includes top level comment, replies are not included.
	CommentCount int

	// RepostCount contains the number of times thread has been reposted.
	RepostCount int

	// FlagCount contains the number of flags that thread has.
	FlagCount int

	// Creator is the account address that created the thread.
	Creator address

	// CreatedAt is thread's creation time as Unix time.
	CreatedAt int64

	// UpdatedAt is thread's update time as unix time.
	UpdatedAt int64
}

// IterateFlags iterates thread moderation flags.
// To reverse iterate use a negative count.
func (t Thread) IterateFlags(start, count int, fn func(boards.Flag) bool) bool {
	return t.ref.Flags.Iterate(start, count, fn)
}

// IterateReposts iterates thread reposts.
// To reverse iterate use a negative count.
func (t Thread) IterateReposts(start, count int, fn func(boardID, repostThreadID uint64) bool) bool {
	return t.ref.Reposts.Iterate(start, count, func(boardID, repostThreadID boards.ID) bool {
		return fn(uint64(boardID), uint64(repostThreadID))
	})
}

// IterateComments iterates thread comments.
// To reverse iterate use a negative count.
func (t Thread) IterateComments(start, count int, fn func(Comment) bool) bool {
	return t.ref.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		return fn(NewSafeComment(comment))
	})
}

// NewSafeThread creates a safe thread.
func NewSafeThread(ref *boards.Post) Thread {
	if !boards.IsThread(ref) {
		panic("post is not a thread")
	}

	return Thread{
		ref:             ref,
		ID:              uint64(ref.ID),
		OriginalBoardID: uint64(ref.OriginalBoardID),
		BoardID:         uint64(ref.Board.ID),
		Title:           ref.Title,
		Body:            ref.Body,
		Hidden:          ref.Hidden,
		Readonly:        ref.Readonly,
		CommentCount:    ref.Replies.Size(),
		RepostCount:     ref.Reposts.Size(),
		FlagCount:       ref.Flags.Size(),
		Creator:         ref.Creator,
		CreatedAt:       timeToUnix(ref.CreatedAt),
		UpdatedAt:       timeToUnix(ref.UpdatedAt),
	}
}
