package boards2

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
)

// TODO: Add unit tests (add interface assert)

// NewReplyStorage creates a new storage for thread replies.
// This is a customized post storage that keeps a flat index of all replies that
// exists within a thread, which allows to find any reply from their thread.
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
	// If post is a direct thread child add it to the post storage
	if p.ParentID == p.ThreadID {
		return s.PostStorage.Add(p)
	}

	// Otherwise when post is a sub-reply add it to the flat index
	k := makeReplyKey(p.ID)
	s.all.Set(k, p)
	return nil
}

func makeReplyKey(id boards.ID) string {
	return id.Key()
}
