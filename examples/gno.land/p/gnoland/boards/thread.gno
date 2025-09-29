package boards

import (
	"errors"
	"std"
	"strings"
	"time"
)

// NewThread creates a new board thread.
func NewThread(b *Board, creator std.Address, title, body string) (*Post, error) {
	if b == nil {
		return nil, errors.New("thread requires a parent board")
	}

	if !creator.IsValid() {
		return nil, errors.New("invalid thread creator address: " + creator.String())
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("thread title is required")
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("thread body is required")
	}

	id := b.generateNextPostID()
	return &Post{
		Title: title,
		Body:  body,

		id:        id,
		threadID:  id,
		board:     b,
		creator:   creator,
		createdAt: time.Now(),
	}, nil
}

// MustNewThread creates a new thread or panics on error.
func MustNewThread(b *Board, creator std.Address, title, body string) *Post {
	t, err := NewThread(b, creator, title, body)
	if err != nil {
		panic(err)
	}
	return t
}

// OriginalBoardID returns the unique identifier of the original thread's board.
// This ID is only available when the thread is a repost.
func (thread Post) OriginalBoardID() ID {
	return thread.originalBoardID
}

// Get returns a reply from a thread.
// No reply is returned if post is not a thread.
func (thread *Post) GetReply(replyID ID) (_ *Post, found bool) {
	if !IsThread(thread) {
		return nil, false
	}

	key := makePostKey(replyID)
	v, found := thread.repliesAll.Get(key)
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

// Repost reposts a thread into another board.
func (thread *Post) Repost(creator std.Address, dst *Board, title, body string) (*Post, error) {
	if !IsThread(thread) {
		return nil, errors.New("post must be a thread to be reposted to another board")
	}

	repost, err := NewThread(dst, creator, title, body)
	if err != nil {
		return nil, err
	}

	repost.parentID = thread.id
	repost.originalBoardID = thread.board.ID()

	// TODO: Ideally we should handle this on Save()
	dst.saveThread(repost)
	thread.saveRepost(repost)

	return repost, nil
}

// DeleteReply deletes a reply from a thread.
func (thread *Post) DeleteReply(replyID ID) error {
	if !IsThread(thread) {
		// TODO: Allow removing replies from parent replies too
		return errors.New("post must be a thread to remove a reply")
	}

	key := makePostKey(replyID)
	v, removed := thread.repliesAll.Remove(key)
	if !removed {
		return errors.New("reply not found")
	}

	// TODO: Shouldn't reply be hidden instead of deleted? Maybe replace reply by a deleted message.
	reply := v.(*Post)
	if reply.ParentID() == thread.id {
		thread.replies.Remove(key)
	} else {
		parent, found := thread.GetReply(reply.ParentID())
		if !found {
			return errors.New("parent reply not found, ID=" + reply.ParentID().String())
		}

		parent.replies.Remove(key)
	}
	return nil
}

func (thread *Post) saveRepost(repost *Post) {
	key := thread.Board().ID().Key()
	thread.reposts.Set(key, repost.ID())
}

// IsThread checks if a post is a thread.
// When a post is not a thread it's considered a thread's reply/comment.
func IsThread(p *Post) bool {
	return p.threadID == p.id
}

// IsRepost checks if a thread is a repost.
func IsRepost(thread *Post) bool {
	return thread.originalBoardID != 0
}
