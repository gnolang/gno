package storage

import (
	"std"

	"gno.land/p/nt/avl"
)

// Storage is general purpose realm storage.
type Storage struct {
	state *avl.Tree
}

// TODO: Introduce sharding later to put an upper bound to AVL tree overhead.

// NewStorage creates a new storage.
func NewStorage() *Storage {
	s := &Storage{
		state: avl.NewTree(),
	}
	return s
}

func (s *Storage) onlyTrustedRealm(state State) {
	if !IsInACL(state, std.PreviousRealm().PkgPath()) {
		panic("unexpected storage access from realm: " + std.PreviousRealm().PkgPath())
	}
}

// State is the minimal interface derived from the AVL interface.
type State interface {
	Get(key string) (value any, exists bool)
	Set(key string, value any) (updated bool)
}

// ExecFn executes caller realm logic on the state.
type ExecFn func(state State) any

// Execute lets a trusted realm execute custom logic on the state (read/write).
func (s *Storage) Execute(accountID string, execFn ExecFn) any {
	tree, ok := s.state.Get(accountID)
	if !ok {
		tree = avl.NewTree()
		s.state.Set(accountID, tree)
	} else {
		s.onlyTrustedRealm(State(tree.(*avl.Tree)))
	}
	return execFn(State(tree.(*avl.Tree)))
}
