package avl

type ITree interface {
	// read operations

	Size() int
	Has(key string) bool
	Get(key string) (value any, exists bool)
	GetByIndex(index int) (key string, value any)
	Iterate(start, end string, cb IterCbFn) bool
	ReverseIterate(start, end string, cb IterCbFn) bool
	IterateByOffset(offset int, count int, cb IterCbFn) bool
	ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool

	// write operations

	Set(key string, value any) (updated bool)
	Remove(key string) (value any, removed bool)
}

type IterCbFn func(key string, value any) bool

//----------------------------------------
// Tree

// The zero struct can be used as an empty tree.
type Tree struct {
	node *Node
}

// NewTree creates a new empty AVL tree.
func NewTree() *Tree {
	return &Tree{
		node: nil,
	}
}

// Size returns the number of key-value pair in the tree.
func (tree *Tree) Size() int {
	return tree.node.Size()
}

// Has checks whether a key exists in the tree.
// It returns true if the key exists, otherwise false.
func (tree *Tree) Has(key string) (has bool) {
	return tree.node.Has(key)
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key exists.
func (tree *Tree) Get(key string) (value any, exists bool) {
	_, value, exists = tree.node.Get(key)
	return
}

// GetByIndex retrieves the key-value pair at the specified index in the tree.
// It returns the key and value at the given index.
func (tree *Tree) GetByIndex(index int) (key string, value any) {
	return tree.node.GetByIndex(index)
}

// Set inserts a key-value pair into the tree.
// If the key already exists, the value will be updated.
// It returns a boolean indicating whether the key was newly inserted or updated.
func (tree *Tree) Set(key string, value any) (updated bool) {
	newnode, updated := tree.node.Set(key, value)
	tree.node = newnode
	return updated
}

// Remove removes a key-value pair from the tree.
// It returns the removed value and a boolean indicating whether the key was found and removed.
func (tree *Tree) Remove(key string) (value any, removed bool) {
	newnode, _, value, removed := tree.node.Remove(key)
	tree.node = newnode
	return value, removed
}

// Iterate performs an in-order traversal of the tree within the specified key range.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) Iterate(start, end string, cb IterCbFn) bool {
	return tree.node.TraverseInRange(start, end, true, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// ReverseIterate performs a reverse in-order traversal of the tree within the specified key range.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) ReverseIterate(start, end string, cb IterCbFn) bool {
	return tree.node.TraverseInRange(start, end, false, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// IterateByOffset performs an in-order traversal of the tree starting from the specified offset.
// It calls the provided callback function for each key-value pair encountered, up to the specified count.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) IterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.node.TraverseByOffset(offset, count, true, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// ReverseIterateByOffset performs a reverse in-order traversal of the tree starting from the specified offset.
// It calls the provided callback function for each key-value pair encountered, up to the specified count.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.node.TraverseByOffset(offset, count, false, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// Verify that Tree implements TreeInterface
var _ ITree = (*Tree)(nil)
