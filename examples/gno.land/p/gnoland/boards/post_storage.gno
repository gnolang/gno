package boards

import (
	"errors"

	"gno.land/p/nt/avl"
)

type (
	// PostIterFn defines a function type to iterate posts.
	PostIterFn func(*Post) bool

	// PostStorage defines an interface for posts storage.
	PostStorage interface {
		// Get retruns a post that matches an ID.
		Get(ID) (_ *Post, found bool)

		// Remove removes a post from the storage.
		Remove(ID) (_ *Post, removed bool)

		// Add adds a post in the storage.
		Add(*Post) error

		// Iterate iterates posts.
		// To reverse iterate posts use a negative count.
		// If the callback returns true, the iteration is stopped.
		Iterate(start, count int, fn PostIterFn) bool
	}
)

// NewPostStorage creates a new storage for posts.
// The new storage uses an AVL tree to store posts.
func NewPostStorage() PostStorage {
	return &postStorage{avl.NewTree()}
}

type postStorage struct {
	posts *avl.Tree // string(Post.ID) -> *Post
}

// Get retruns a post that matches an ID.
func (s postStorage) Get(id ID) (*Post, bool) {
	key := makePostKey(id)
	v, found := s.posts.Get(key)
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

// Remove removes a post from the storage.
func (s *postStorage) Remove(id ID) (*Post, bool) {
	key := makePostKey(id)
	v, removed := s.posts.Remove(key)
	if !removed {
		return nil, false
	}
	return v.(*Post), true
}

// Add adds a post in the storage.
// It updates existing posts when storage contains one with the same ID.
func (s *postStorage) Add(p *Post) error {
	if p == nil {
		return errors.New("saving a nil post is not supported")
	}

	s.posts.Set(makePostKey(p.ID), p)
	return nil
}

// Iterate iterates posts.
// To reverse iterate posts use a negative count.
// If the callback returns true, the iteration is stopped.
func (s postStorage) Iterate(start, count int, fn PostIterFn) bool {
	if count < 0 {
		return s.posts.ReverseIterateByOffset(start, -count, func(_ string, v any) bool {
			return fn(v.(*Post))
		})
	}

	return s.posts.IterateByOffset(start, count, func(_ string, v any) bool {
		return fn(v.(*Post))
	})
}

func makePostKey(postID ID) string {
	return postID.PaddedString()
}
