package bptree

import (
	"fmt"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
)

// Tree is the interface for both mutable and immutable B+ trees.
// Mirrors the iavl store's Tree interface but uses bptree types.
type Tree interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) (bool, error)
	Remove(key []byte) ([]byte, bool, error)
	SaveVersion() ([]byte, int64, error)
	DeleteVersionsTo(version int64) error
	Version() int64
	Size() int64
	Hash() []byte
	GetLatestVersion() (int64, error)
	VersionExists(version int64) bool
	GetVersioned(key []byte, version int64) ([]byte, error)
	GetImmutableTree(version int64) (*bp.ImmutableTree, error)
}

// Verify MutableTree implements Tree.
var _ Tree = (*mutableTreeAdapter)(nil)

// mutableTreeAdapter wraps bp.MutableTree to implement Tree.
type mutableTreeAdapter struct {
	*bp.MutableTree
}

func (a *mutableTreeAdapter) GetLatestVersion() (int64, error) {
	return a.Version(), nil
}

func (a *mutableTreeAdapter) GetImmutableTree(version int64) (*bp.ImmutableTree, error) {
	return a.MutableTree.GetImmutable(version)
}

// immutableTreeAdapter wraps bp.ImmutableTree to implement Tree.
// Mutations panic.
type immutableTreeAdapter struct {
	*bp.ImmutableTree
}

func (a *immutableTreeAdapter) Set(_, _ []byte) (bool, error) {
	panic("cannot Set on immutable B+ tree")
}

func (a *immutableTreeAdapter) Remove(_ []byte) ([]byte, bool, error) {
	panic("cannot Remove on immutable B+ tree")
}

func (a *immutableTreeAdapter) SaveVersion() ([]byte, int64, error) {
	panic("cannot SaveVersion on immutable B+ tree")
}

func (a *immutableTreeAdapter) DeleteVersionsTo(_ int64) error {
	panic("cannot DeleteVersionsTo on immutable B+ tree")
}

func (a *immutableTreeAdapter) GetLatestVersion() (int64, error) {
	return a.Version(), nil
}

func (a *immutableTreeAdapter) VersionExists(version int64) bool {
	return a.Version() == version
}

func (a *immutableTreeAdapter) GetVersioned(key []byte, version int64) ([]byte, error) {
	if a.Version() != version {
		return nil, nil
	}
	return a.Get(key)
}

func (a *immutableTreeAdapter) GetImmutableTree(version int64) (*bp.ImmutableTree, error) {
	if a.Version() != version {
		return nil, fmt.Errorf("version mismatch: got %d, want %d", version, a.Version())
	}
	return a.ImmutableTree, nil
}

var _ Tree = (*immutableTreeAdapter)(nil)
