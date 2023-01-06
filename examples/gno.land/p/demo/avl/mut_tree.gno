package avl

// Tree

// The zero struct can be used as an empty tree.
type Tree struct {
	node *Node
}

func NewTree() *Tree {
	return &Tree{
		node: nil,
	}
}

func (tree *Tree) Size() int {
	return tree.node.Size()
}

func (tree *Tree) Has(key string) (has bool) {
	return tree.node.Has(key)
}

func (tree *Tree) Get(key string) (value interface{}, exists bool) {
	_, value, exists = tree.node.Get(key)
	return
}

func (tree *Tree) GetByIndex(index int) (key string, value interface{}) {
	return tree.node.GetByIndex(index)
}

func (tree *Tree) Set(key string, value interface{}) (updated bool) {
	newnode, updated := tree.node.Set(key, value)
	node.node = newnode
	return updated
}

func (tree *Tree) Remove(key string) (value interface{}, removed bool) {
	newnode, _, value, removed := tree.node.Remove(key)
	tree.node = newnode
	return value, removed
}

// Shortcut for TraverseInRange.
func (tree *Tree) Iterate(start, end string, cb func(*Node) bool) bool {
	return tree.node.TraverseInRange(start, end, true, true, cb)
}

// Shortcut for TraverseInRange.
func (tree *Tree) IterateReverse(start, end string, cb func(*Node) bool) bool {
	return tree.node.TraverseInRange(start, end, false, true, cb)
}
