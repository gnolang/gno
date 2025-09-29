package boards

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/nt/avl"
)

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a board that holds other replies.
type Post struct {
	Title     string
	Body      string
	Hidden    bool
	Readonly  bool
	UpdatedAt time.Time

	id              ID
	parentID        ID // Parent Post.ID for all replies or reposted thread
	threadID        ID // Post.ID of the thread where reply is created
	originalBoardID ID // Original Board.ID when thread is a repost
	board           *Board
	flags           avl.Tree // std.Address -> string(reason)
	replies         avl.Tree // string(Post.ID) -> *Post (direct post replies)
	repliesAll      avl.Tree // string(Post.ID) -> *Post (all thread replies)
	reposts         avl.Tree // string(Board.ID) -> Post.ID
	creator         std.Address
	createdAt       time.Time
}

// ID returns the unique identifier of the post.
func (post Post) ID() ID {
	return post.id
}

// ParentID returns the unique identifier of the post's parent.
func (post Post) ParentID() ID {
	return post.parentID
}

// Board returns the board where post was created.
func (post Post) Board() *Board {
	return post.board
}

// Creator returns the address of the account that created the post.
func (post Post) Creator() std.Address {
	return post.creator
}

// CreatedAt returns the time when post was created.
func (post Post) CreatedAt() time.Time {
	return post.createdAt
}

// Summary return a summary of the post's body.
// It returns the body making sure that the length is limited to 80 characters.
func (post Post) Summary() string {
	return summaryOf(post.Body, 80)
}

// Flag add a flag to the post.
// It returns false when the user flagging the post already flagged it.
func (post *Post) Flag(user std.Address, reason string) bool {
	if post.flags.Has(user.String()) {
		return false
	}

	post.flags.Set(user.String(), reason)
	return true
}

// FlagsCount returns the number of times post was flagged.
func (post Post) FlagsCount() int {
	return post.flags.Size()
}

// HasReplies checks if a thread or reply has replies.
func (post Post) HasReplies() bool {
	return post.replies.Size() > 0
}

// Save stores the thread within it's parent board, or if post
// is a reply it stores the reply within it's parent thread.
func (post *Post) Save() error {
	if IsThread(post) {
		post.board.saveThread(post)
	} else {
		thread, found := post.board.GetThread(post.threadID)
		if !found {
			return errors.New("reply's thread not found, ID=" + post.threadID.String())
		}

		// Add reply to the flat replies index. This index allows listing all thread
		// replies without needing to recursively explore the child replies of each reply.
		key := makePostKey(post.id)
		thread.repliesAll.Set(key, post)

		// Add reply to it's parent index.
		// Reply parent can be a thread or another reply.
		parent := thread
		if post.parentID != thread.ID() {
			parentKey := makePostKey(post.parentID)
			v, found := thread.repliesAll.Get(parentKey)
			if !found {
				return errors.New("reply's parent post not found, ID=" + post.parentID.String())
			}

			parent = v.(*Post)
		}
		parent.replies.Set(key, post)
	}
	return nil
}

func makePostKey(postID ID) string {
	return postID.PaddedString()
}

func summaryOf(text string, length int) string {
	lines := strings.SplitN(text, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		line = line + "..."
	}
	return line
}
