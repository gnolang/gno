package boards

import (
	"errors"
	"strings"

	"gno.land/p/nt/avl"
)

type (
	// BoardIterFn defines a function type to iterate boards.
	BoardIterFn func(*Board) bool

	// Storage defines an interface for boards storage.
	Storage interface {
		// Get retruns a boards that matches an ID.
		Get(ID) (_ *Board, found bool)

		// GetByName retruns a boards that matches a name.
		GetByName(name string) (_ *Board, found bool)

		// Remove removes a board from the storage.
		Remove(ID) (_ *Board, removed bool)

		// Add adds a board to the storage.
		Add(*Board) error

		// Size returns the number of boards in the storage.
		Size() int

		// Iterate iterates boards.
		// To reverse iterate boards use a negative count.
		// If the callback returns true, the iteration is stopped.
		Iterate(start, count int, fn BoardIterFn) bool
	}
)

// NewStorage creates a new boards storage.
func NewStorage() Storage {
	return &storage{
		byID:   avl.NewTree(),
		byName: avl.NewTree(),
	}
}

type storage struct {
	byID   *avl.Tree // string(Board.ID) -> *Board
	byName *avl.Tree // Board.Name -> Board.ID
}

// Get returns a board for a specific ID.
func (s storage) Get(boardID ID) (*Board, bool) {
	key := makeBoardKey(boardID)
	v, found := s.byID.Get(key)
	if !found {
		return nil, false
	}
	return v.(*Board), true
}

// Get returns a board for a specific name.
func (s storage) GetByName(name string) (*Board, bool) {
	key := makeBoardNameKey(name)
	v, found := s.byName.Get(key)
	if !found {
		return nil, false
	}
	return s.Get(v.(ID))
}

// Remove removes a board from the storage.
// It returns false when board is not found.
func (s *storage) Remove(boardID ID) (*Board, bool) {
	board, found := s.Get(boardID)
	if !found {
		return nil, false
	}

	// Remove indexes for current and previous board names
	names := append([]string{board.Name}, board.Aliases...)
	for _, name := range names {
		key := makeBoardNameKey(name)

		// Make sure that name is indexed to the board being removed
		v, found := s.byName.Get(key)
		if found && v.(ID) == boardID {
			s.byName.Remove(key)
		}
	}

	key := makeBoardKey(board.ID)
	_, removed := s.byID.Remove(key)
	return board, removed
}

// Add adds a board to the storage.
// If board already exists it updates storage by reindexing the board by ID and name.
// When board name changes it's indexed so it can be found with the new and previous names.
func (s *storage) Add(board *Board) error {
	if board == nil {
		return errors.New("adding nil boards to the storage is not allowed")
	}

	key := makeBoardKey(board.ID)
	s.byID.Set(key, board)

	// Index by name when the optional board name is not empty
	if key = makeBoardNameKey(board.Name); key != "" {
		s.byName.Set(key, board.ID)
	}
	return nil
}

// Size returns the number of boards in the storage.
func (s storage) Size() int {
	return s.byID.Size()
}

// Iterate iterates boards.
// To reverse iterate boards use a negative count.
// If the callback returns true, the iteration is stopped.
func (s storage) Iterate(start, count int, fn BoardIterFn) bool {
	if count < 0 {
		return s.byID.ReverseIterateByOffset(start, -count, func(_ string, v any) bool {
			return fn(v.(*Board))
		})
	}

	return s.byID.IterateByOffset(start, count, func(_ string, v any) bool {
		return fn(v.(*Board))
	})
}

func makeBoardKey(boardID ID) string {
	return boardID.Key()
}

func makeBoardNameKey(name string) string {
	name = strings.TrimSpace(name)
	return strings.ToLower(name)
}
