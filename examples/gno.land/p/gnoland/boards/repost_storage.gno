package boards

import (
	"errors"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

type (
	// RepostIterFn defines a function type to iterate reposts.
	RepostIterFn func(board, repost ID) bool

	// RepostStorage defines an interface for storing reposts.
	RepostStorage interface {
		// Get returns the repost ID for a board.
		Get(board ID) (repost ID, found bool)

		// Add adds a new repost to the storage.
		Add(repost *Post) error

		// Remove removes repost for a board.
		Remove(board ID) (removed bool)

		// Size returns the number of reposts in the storage.
		Size() int

		// Iterate iterates reposts.
		// To reverse iterate reposts use a negative count.
		// If the callback returns true, the iteration is stopped.
		Iterate(start, count int, fn RepostIterFn) bool
	}
)

// NewRepostStorage creates a new storage for reposts.
// The new storage uses an AVL tree to store reposts.
func NewRepostStorage() RepostStorage {
	return &repostStorage{avl.NewTree()}
}

type repostStorage struct {
	reposts *avl.Tree // string(Board.ID) -> Post.ID
}

// Get returns the repost ID for a board.
func (s repostStorage) Get(boardID ID) (ID, bool) {
	v, found := s.reposts.Get(boardID.Key())
	if !found {
		return 0, false
	}
	return v.(ID), true
}

// Add adds a new repost to the storage.
func (s *repostStorage) Add(repost *Post) error {
	if repost == nil {
		return errors.New("saving nil reposts is not allowed")
	}

	s.reposts.Set(repost.Board.ID.Key(), repost.ID)
	return nil
}

// Remove removes repost for a board.
func (s *repostStorage) Remove(boardID ID) bool {
	_, removed := s.reposts.Remove(boardID.Key())
	return removed
}

// Size returns the number of reposts in the storage.
func (s repostStorage) Size() int {
	return s.reposts.Size()
}

// Iterate iterates reposts.
// To reverse iterate reposts use a negative count.
// If the callback returns true, the iteration is stopped.
func (s repostStorage) Iterate(start, count int, fn RepostIterFn) bool {
	if count < 0 {
		return s.reposts.ReverseIterateByOffset(start, -count, func(k string, v any) bool {
			id, err := seqid.FromString(k)
			if err != nil {
				panic(err)
			}

			return fn(ID(id), v.(ID))
		})
	}

	return s.reposts.IterateByOffset(start, count, func(k string, v any) bool {
		id, err := seqid.FromString(k)
		if err != nil {
			panic(err)
		}

		return fn(ID(id), v.(ID))
	})
}
