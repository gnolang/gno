package hub

import (
	"gno.land/p/gnoland/boards"
)

// Flag defines a type for thread and comment flags.
type Flag struct {
	// user is the user that flagged.
	user address

	// reason is the reason for flagging.
	reason string
}

// User returns the user that flagged.
func (f Flag) User() address { return f.user }

// Reason returns the reason for flagging.
func (f Flag) Reason() string { return f.reason }

// NewSafeFlag creates a safe flag.
func NewSafeFlag(ref boards.Flag) Flag {
	return Flag{
		user:   ref.User,
		reason: ref.Reason,
	}
}

// Thread defines a type for board threads.
type Thread struct {
	// id is the unique identifier of the thread.
	id uint64

	// originalBoardID contains the board ID of the original thread when current is a repost.
	originalBoardID uint64

	// originalThreadID contains the ID of the original thread when current is a repost.
	originalThreadID uint64

	// boardID is the board ID where thread is created.
	boardID uint64

	// title contains thread's title.
	title string

	// body contains content of the thread.
	body string

	// hidden indicates that thread is hidden.
	hidden bool

	// readonly indicates that thread is readonly.
	readonly bool

	// commentCount contains the number of thread comments.
	// Count only includes top level comment, replies are not included.
	commentCount int

	// repostCount contains the number of times thread has been reposted.
	repostCount int

	// flagCount contains the number of flags that thread has.
	flagCount int

	// creator is the account address that created the thread.
	creator address

	// createdAt is thread's creation time as Unix time.
	createdAt int64

	// updatedAt is thread's update time as unix time.
	updatedAt int64
}

// ID returns the unique identifier of the thread.
func (t Thread) ID() uint64 { return t.id }

// OriginalBoardID returns the board ID of the original thread when current is a repost.
func (t Thread) OriginalBoardID() uint64 { return t.originalBoardID }

// OriginalThreadID returns the ID of the original thread when current is a repost.
func (t Thread) OriginalThreadID() uint64 { return t.originalThreadID }

// BoardID returns the board ID where the thread is created.
func (t Thread) BoardID() uint64 { return t.boardID }

// Title returns the thread's title.
func (t Thread) Title() string { return t.title }

// Body returns the content of the thread.
func (t Thread) Body() string { return t.body }

// Hidden indicates that the thread is hidden.
func (t Thread) Hidden() bool { return t.hidden }

// Readonly indicates that the thread is readonly.
func (t Thread) Readonly() bool { return t.readonly }

// CommentCount returns the number of thread comments.
// Count only includes top level comment, replies are not included.
func (t Thread) CommentCount() int { return t.commentCount }

// RepostCount returns the number of times the thread has been reposted.
func (t Thread) RepostCount() int { return t.repostCount }

// FlagCount returns the number of flags that the thread has.
func (t Thread) FlagCount() int { return t.flagCount }

// Creator returns the account address that created the thread.
func (t Thread) Creator() address { return t.creator }

// CreatedAt returns the thread's creation time as Unix time.
func (t Thread) CreatedAt() int64 { return t.createdAt }

// UpdatedAt returns the thread's update time as Unix time.
func (t Thread) UpdatedAt() int64 { return t.updatedAt }

// NewSafeThread creates a safe thread.
func NewSafeThread(ref *boards.Post) Thread {
	if !boards.IsThread(ref) {
		panic("post is not a thread")
	}

	var commentCount int
	if ref.Replies != nil {
		commentCount = ref.Replies.Size()
	}

	var repostCount int
	if ref.Reposts != nil {
		repostCount = ref.Reposts.Size()
	}

	var flagCount int
	if ref.Flags != nil {
		flagCount = ref.Flags.Size()
	}

	return Thread{
		id:               uint64(ref.ID),
		originalBoardID:  uint64(ref.OriginalBoardID),
		originalThreadID: uint64(ref.ParentID),
		boardID:          uint64(ref.Board.ID),
		title:            ref.Title,
		body:             ref.Body,
		hidden:           ref.Hidden,
		readonly:         ref.Readonly,
		commentCount:     commentCount,
		repostCount:      repostCount,
		flagCount:        flagCount,
		creator:          ref.Creator,
		createdAt:        timeToUnix(ref.CreatedAt),
		updatedAt:        timeToUnix(ref.UpdatedAt),
	}
}
