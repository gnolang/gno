package boards2

import (
	"errors"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
)

// NewReplyStorage creates a new storage for thread replies.
// This is a customized post storage that also keeps a flat index with all replies
// that exists within a thread, which allows to get any reply from their thread.
func NewReplyStorage() boards.PostStorage {
	return &replyStorage{
		PostStorage: boards.NewPostStorage(),
		all:         avl.NewTree(),
	}
}

type replyStorage struct {
	boards.PostStorage

	// Flat index to store all replies that exists within a thread
	all *avl.Tree // string(Post.ID) -> *Post
}

// Get retruns a reply that matches an ID.
// Reply can be a direct thread reply or a sub-reply.
func (s replyStorage) Get(id boards.ID) (*boards.Post, bool) {
	post, found := s.PostStorage.Get(id)
	if found {
		return post, true
	}

	k := makeReplyKey(id)
	v, found := s.all.Get(k)
	if !found {
		return nil, false
	}
	return v.(*boards.Post), true
}

// Remove removes a post from the storage.
func (s *replyStorage) Remove(id boards.ID) (*boards.Post, bool) {
	post, removed := s.PostStorage.Remove(id)
	if removed {
		return post, true
	}

	// When reply is not a direct thread reply try to remove it from the flat index
	k := makeReplyKey(id)
	v, removed := s.all.Remove(k)
	if !removed {
		return nil, false
	}
	return v.(*boards.Post), true
}

// Add adds a post in the storage.
// It updates existing posts when storage contains one with the same ID.
func (s *replyStorage) Add(p *boards.Post) error {
	if p == nil {
		return errors.New("saving nil replies is not allowed")
	}

	// If post is a direct thread child add it to the post storage
	if p.ParentID == p.ThreadID {
		return s.PostStorage.Add(p)
	}

	// Otherwise when post is a sub-reply add it to the flat index
	k := makeReplyKey(p.ID)
	s.all.Set(k, p)
	return nil
}

// Size returns the number of direct replies in the storage.
// It doesn't includes sub-replies.
func (s replyStorage) Size() int {
	return s.PostStorage.Size()
}

// Iterate iterates direct replies.
// To reverse iterate posts use a negative count.
// If the callback returns true, the iteration is stopped.
// Sub-replies are NOT iterated.
func (s replyStorage) Iterate(start, count int, fn boards.PostIterFn) bool {
	return s.PostStorage.Iterate(start, count, fn)
}

func makeReplyKey(id boards.ID) string {
	return id.Key()
}
