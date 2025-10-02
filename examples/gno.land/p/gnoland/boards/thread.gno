package boards

import (
	"errors"
	"strings"
	"time"
)

// NewThread creates a new board thread.
func NewThread(b *Board, creator address, title, body string) (*Post, error) {
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
func MustNewThread(b *Board, creator address, title, body string) *Post {
	t, err := NewThread(b, creator, title, body)
	if err != nil {
		panic(err)
	}
	return t
}

// NewRepost creates a new thread that is a repost of a thread from another board.
func NewRepost(thread *Post, dst *Board, creator address) (*Post, error) {
	if thread == nil {
		return nil, errors.New("thread to repost is required")
	}

	if !IsThread(thread) {
		return nil, errors.New("post must be a thread to be reposted to another board")
	}

	if IsRepost(thread) {
		return nil, errors.New("reposting a thread that is a repost is not allowed")
	}

	if dst == nil {
		return nil, errors.New("thread repost requires a destination board")
	}

	if !creator.IsValid() {
		return nil, errors.New("invalid thread repost creator address: " + creator.String())
	}

	id := dst.generateNextPostID()
	return &Post{
		id:              id,
		threadID:        id,
		board:           dst,
		parentID:        thread.id,
		originalBoardID: thread.board.ID(),
		creator:         creator,
		createdAt:       time.Now(),
	}, nil
}

// MustNewRepost creates a new thread that is a repost of a thread from another board or panics on error.
func MustNewRepost(thread *Post, dst *Board, creator address) *Post {
	r, err := NewRepost(thread, dst, creator)
	if err != nil {
		panic(err)
	}
	return r
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
