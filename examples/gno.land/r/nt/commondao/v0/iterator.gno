package commondao

import (
	"chain/runtime"

	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/commondao/v0"
)

// NewIterator returns a SubDAO iterator.
func NewIterator(rootID uint64, options ...IteratorOption) Iterator {
	assertIsOwner(runtime.CurrentRealm().Address(), rootID)

	tree := getTree(rootID)
	if tree == nil {
		return Iterator{}
	}

	it := Iterator{tree: tree}
	for _, apply := range options {
		apply(&it)
	}

	size := it.tree.Size()
	if it.count > 0 {
		it.maxIndex = it.current + it.count
		if it.maxIndex > size {
			it.maxIndex = size
		}
	} else {
		it.maxIndex = size
	}

	it.count = it.maxIndex - it.current

	return it
}

// Iterator defines an iterator of SubDAOs.
type Iterator struct {
	tree                     rotree.IReadOnlyTree
	count, current, maxIndex int
	path                     string
	dao                      *commondao.CommonDAO
}

// Count returns the number of DAOs to iterate.
func (it Iterator) Count() int {
	return it.count
}

// Next returns true when a new DAO is available.
func (it *Iterator) Next() bool {
	if it.tree == nil || it.current == it.maxIndex {
		return false
	}

	path, v := it.tree.GetByIndex(it.current)
	if v == nil {
		return false
	}

	it.dao = v.(*commondao.CommonDAO)
	it.path = path
	it.current++
	return true
}

// DAO returns the current DAO.
func (it Iterator) DAO() *commondao.CommonDAO {
	return it.dao
}

// Path returns the current DAO path.
func (it Iterator) Path() string {
	return it.path
}
