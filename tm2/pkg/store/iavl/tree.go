package iavl

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/iavl"
)

var (
	_ Tree = (*immutableTree)(nil)
	_ Tree = (*iavl.MutableTree)(nil)
)

// Tree defines an interface that both mutable and immutable IAVL trees
// must implement. For mutable IAVL trees, the interface is directly
// implemented by an iavl.MutableTree. For an immutable IAVL tree, a wrapper
// must be made.
type Tree interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) (bool, error)
	Remove(key []byte) ([]byte, bool, error)
	SaveVersion() ([]byte, int64, error)
	DeleteVersionsTo(version int64) error
	Version() int64
	Hash() []byte
	GetLatestVersion() (int64, error)
	VersionExists(version int64) bool
	GetVersioned(key []byte, version int64) ([]byte, error)
	GetImmutable(version int64) (*iavl.ImmutableTree, error)
}

// immutableTree is a simple wrapper around a reference to an iavl.ImmutableTree
// that implements the Tree interface. It should only be used for querying
// and iteration, specifically at previous heights.
type immutableTree struct {
	*iavl.ImmutableTree
}

func (it *immutableTree) Set(_, _ []byte) (bool, error) {
	panic("cannot call 'Set' on an immutable IAVL tree")
}

func (it *immutableTree) Remove(_ []byte) ([]byte, bool, error) {
	panic("cannot call 'Remove' on an immutable IAVL tree")
}

func (it *immutableTree) SaveVersion() ([]byte, int64, error) {
	panic("cannot call 'SaveVersion' on an immutable IAVL tree")
}

func (it *immutableTree) DeleteVersionsTo(_ int64) error {
	panic("cannot call 'DeleteVersionsTo' on an immutable IAVL tree")
}

func (it *immutableTree) GetLatestVersion() (int64, error) {
	return it.Version(), nil
}

func (it *immutableTree) VersionExists(version int64) bool {
	return it.Version() == version
}

func (it *immutableTree) GetVersioned(key []byte, version int64) ([]byte, error) {
	if it.Version() != version {
		return nil, nil
	}

	return it.Get(key)
}

func (it *immutableTree) GetImmutable(version int64) (*iavl.ImmutableTree, error) {
	if it.Version() != version {
		return nil, fmt.Errorf("version mismatch on immutable IAVL tree; got: %d, expected: %d", version, it.Version())
	}

	return it.ImmutableTree, nil
}
