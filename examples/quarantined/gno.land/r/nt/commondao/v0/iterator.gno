package commondao

import (
	"gno.land/p/nt/bptree/v0/rotree"
	"gno.land/p/nt/commondao/v0"
)

// NewIterator returns a SubDAO iterator.
func NewIterator(cur realm, rootID uint64, options ...IteratorOption) Iterator {
	assertIsOwner(cur.Previous().Address(), rootID)

	tree := getTree(rootID)
	if tree == nil {
		return Iterator{}
	}

	it := Iterator{tree: tree}
	for _, apply := range options {
		apply(&it)
	}

	size := it.tree.Size()
	if it.current > size {
		it.current = size
	}

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

	_, v := it.tree.GetByIndex(it.current)
	if v == nil {
		return false
	}

	it.dao = v.(*commondao.CommonDAO)
	it.current++
	return true
}

// DAO returns a read only view of the current DAO.
func (it Iterator) DAO() commondao.ReadonlyCommonDAO {
	return it.dao.Readonly()
}
