package avl

// MutTree

type MutTree struct {
	tree *Tree
}

func NewMutTree() *MutTree {
	return &MutTree{
		tree: nil,
	}
}

func (mtree *MutTree) Size() int {
	return mtree.tree.Size()
}

func (mtree *MutTree) Has(key string) (has bool) {
	return mtree.tree.Has(key)
}

func (mtree *MutTree) Get(key string) (value interface{}, exists bool) {
	_, value, exists = mtree.tree.Get(key)
	return
}

func (mtree *MutTree) GetByIndex(index int) (key string, value interface{}) {
	return mtree.tree.GetByIndex(index)
}

func (mtree *MutTree) Set(key string, value interface{}) (updated bool) {
	newtree, updated := mtree.tree.Set(key, value)
	mtree.tree = newtree
	return updated
}

func (mtree *MutTree) Remove(key string) (value interface{}, removed bool) {
	newtree, _, value, removed := mtree.tree.Remove(key)
	mtree.tree = newtree
	return value, removed
}

// Shortcut for TraverseInRange.
func (mtree *MutTree) Iterate(start, end string, cb func(*Tree) bool) bool {
	return mtree.tree.TraverseInRange(start, end, true, true, cb)
}

// Shortcut for TraverseInRange.
func (mtree *MutTree) IterateReverse(start, end string, cb func(*Tree) bool) bool {
	return mtree.tree.TraverseInRange(start, end, false, true, cb)
}
