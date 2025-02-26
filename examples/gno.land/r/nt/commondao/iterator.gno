package commondao

import (
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/nt/commondao"
)

func newIterator(tree rotree.IReadOnlyTree, options ...IteratorOption) Iterator {
	if tree == nil {
		return Iterator{}
	}

	it := Iterator{tree: tree}
	for _, apply := range options {
		apply(&it)
	}

	if it.count > 0 {
		it.maxIndex = it.current + it.count
	} else {
		it.maxIndex = it.tree.Size()
		it.count = it.tree.Size()
	}

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
	if v != nil {
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
