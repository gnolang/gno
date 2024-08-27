package btree

import (
	"errors"
)

// Struct to store the information required
type Content struct {
	Key   int
	Value interface{}
}

// Nodes to use for the BTree
type BTreeNode struct {
	keys     []Content    // keys in node
	children []*BTreeNode // pointers to children
	leaf     bool         // Node is leaf or not
	n        int          // amount of stored vals in node
}

// BTree struct
type BTree struct {
	root *BTreeNode
	t    int // min grade
	size int
}

// Create new B Tree
func NewBTree(t int) (*BTree, error) {
	// If t == 1, then it's a binary tree, so we skip this case.
	if t < 2 {
		return nil, errors.New("invalid t")
	}
	return &BTree{
		root: &BTreeNode{
			keys:     make([]Content, 0),    // Create empty array since it's just a new node
			children: make([]*BTreeNode, 0), // Create empty array since no node has been added yet
			leaf:     true,
		},
		t: t,
	}, nil
}

// Search in B-tree
func (bt *BTree) Search(k int) (Content, error) {
	return search(bt.root, k)
}

func search(bt *BTreeNode, k int) (Content, error) {
	i := 0
	// Checking the keys in the current node until reaching one bigger key than the one we're looking for
	for i < bt.n && k > bt.keys[i].Key {
		i++
	}
	// Verify if the idx is less than the amount of available keys, and if the node with this index is
	// equal to the key we're looking for, we return it
	if i < bt.n && k == bt.keys[i].Key {
		return bt.keys[i], nil
		// If not, and the node is a leaf this means we can't advance, so we return it doesn't exist
	} else if bt.leaf {
		return Content{}, errors.New("Value not existing")
		// If it's not a leaf, we can continue forward
	} else {
		return search(bt.children[i], k)
	}
}

func (bt *BTree) splitChild(parent *BTreeNode, childIndex int) {
	t := bt.t // grade of btree

	// Child node to be divided
	oldChild := parent.children[childIndex]

	// New node to hold remaining keys and children
	newChild := &BTreeNode{
		keys:     make([]Content, t-1),  // t-1 spaces to fill
		children: make([]*BTreeNode, 0), // no children
		leaf:     oldChild.leaf,         // since this node will inherit oldChild position, must be the sabe as previous
		n:        t - 1,                 // len(keys)
	}

	// Copies last (t-1) keys from oldChild to newChild
	copy(newChild.keys, oldChild.keys[t:])

	// If was a leaf, just copy the children too
	if !oldChild.leaf {
		newChild.children = make([]*BTreeNode, t) // But will require to update the number of children
		copy(newChild.children, oldChild.children[t:])
	}

	// Middle value that goes up
	middle := oldChild.keys[t-1]

	// Adjusting original node to hold only the first t-1 keys
	oldChild.keys = oldChild.keys[:t-1]
	oldChild.n = t - 1 // len(oldChild.keys) updated

	// Inserting new node in the children array of parent
	parent.children = append(parent.children[:childIndex+1], append([]*BTreeNode{newChild}, parent.children[childIndex+1:]...)...)

	// Moving the central key of the original node into the parent
	parent.keys = append(parent.keys[:childIndex], append([]Content{middle}, parent.keys[childIndex:]...)...)

	// Updating parent size
	parent.n++
}

// Insert new val
func (bt *BTree) Insert(k Content) {
	root := bt.root
	// Follow the rule of b-trees and split it if
	// surpasses the 2*grade size
	if root.n == 2*bt.t-1 {
		s := &BTreeNode{
			keys:     make([]Content, 0),
			children: make([]*BTreeNode, 0),
			leaf:     false,
		}
		bt.root = s
		s.children = append(s.children, root)
		// Splitting to follow the rule
		bt.splitChild(s, 0)
		// Adding the node k in s
		bt.insertNonFull(s, k)
	} else {
		// Otherwise just add the key in the root
		bt.insertNonFull(root, k)
	}
	// This is to get the size without counting each time
	bt.size++
}

// Insert if the node isn't full
func (bt *BTree) insertNonFull(x *BTreeNode, k Content) {
	i := x.n - 1 // Index for new item
	if x.leaf {
		x.keys = append(x.keys, Content{Key: 0}) // If it's leaf, just add the val in the end
		for i >= 0 && k.Key < x.keys[i].Key {
			x.keys[i+1] = x.keys[i]
			i--
		}
		x.keys[i+1] = k
		x.n++
	} else {
		for i >= 0 && k.Key < x.keys[i].Key {
			i--
		}
		i++
		if x.children[i].n == 2*bt.t-1 {
			bt.splitChild(x, i)
			if k.Key > x.keys[i].Key {
				i++
			}
		}
		bt.insertNonFull(x.children[i], k)
	}
}

// Delete key
func (bt *BTree) Delete(k int) {
	if bt.root == nil || bt.GetSize() == 0 {
		return
	}
	removed := bt.delete(bt.root, k)
	if bt.root.n == 0 {
		if bt.root.leaf {
			bt.root = nil
		} else {
			bt.root = bt.root.children[0]
		}
	}
	if removed {
		bt.size--
	}
}

func (bt *BTree) delete(x *BTreeNode, k int) bool {
	removed := false
	t := bt.t
	i := 0

	// Moving forward until reaching a key bigger than target
	for i < x.n && k > x.keys[i].Key {
		i++
	}

	// Case we find the key
	if i < x.n && k == x.keys[i].Key {
		if x.leaf {
			// If we're in a leaf, just delete the key
			x.keys = append(x.keys[:i], x.keys[i+1:]...)
			x.n--
			removed = true
		} else {
			// If it's inside an internal node
			if len(x.children[i].keys) >= t {
				// Get the predecessor
				pred := x.children[i]
				// Dive until reaching the rightmost leaf
				for !pred.leaf {
					pred = pred.children[pred.n]
				}
				// Get the last key of the leaf and replace it
				predKey := pred.keys[pred.n-1]
				x.keys[i] = predKey
				// Remove the key from the leaf
				removed = bt.delete(x.children[i], predKey.Key)
			} else if len(x.children[i+1].keys) >= t {
				// Get the successor
				succ := x.children[i+1]
				// Dive until reaching the leftmost leaf
				for !succ.leaf {
					succ = succ.children[0]
				}
				// Get the first key of the leaf and replace it
				succKey := succ.keys[0]
				x.keys[i] = succKey
				// Remove the key from the leaf
				removed = bt.delete(x.children[i+1], succKey.Key)
			} else {
				// If both predecessor and successor don't have enough keys
				// merge the key with the right child
				bt.merge(x, i)
				// After merging, we need to delete from the merged child
				removed = bt.delete(x.children[i], k)
			}
		}
	} else if !x.leaf {
		// Dive deeper since the key isn't here
		if len(x.children[i].keys) < t {
			// Fill the child if it has less than t keys
			bt.fill(x, i)
		}
		// Recursively delete from the appropriate child
		if i > x.n {
			// If the key is in the last child, we need to go to the previous one
			removed = bt.delete(x.children[i-1], k)
		} else {
			// Otherwise, we just go to the child
			removed = bt.delete(x.children[i], k)
		}
	}
	return removed
}

// Fills child x.children[i] if it has less than t keys
func (bt *BTree) fill(x *BTreeNode, i int) {
	t := bt.t // grade of btree

	// Check if the left sibling has enough keys to borrow
	if i != 0 && len(x.children[i-1].keys) >= t {
		bt.borrowFromPrev(x, i)

		// Check if the right sibling has enough keys to borrow
	} else if i != x.n && len(x.children[i+1].keys) >= t {
		bt.borrowFromNext(x, i)

		// If neither sibling has enough keys, merge with a sibling
	} else {
		if i != x.n {
			// Merging with next sibling because it's not the last child
			bt.merge(x, i)
		} else {
			// Merging with previous sibling because it's the last child
			bt.merge(x, i-1)
		}
	}
}

// Borrows a key from node x.children[i-1] and includes it in x.children[i]
func (bt *BTree) borrowFromPrev(x *BTreeNode, i int) {
	// Get the child and sibling
	child := x.children[i]
	sibling := x.children[i-1]

	// Shift all keys in the child to the right
	for j := child.n - 1; j >= 0; j-- {
		child.keys[j+1] = child.keys[j]
	}
	// If the child is not a leaf, shift all children in the child to the right
	if !child.leaf {
		for j := child.n; j >= 0; j-- {
			child.children[j+1] = child.children[j]
		}
	}
	// Move the key from the parent to the child. Remember this keeps the order because
	// the key is the one that was borrowed from the sibling.
	child.keys[0] = x.keys[i-1]
	// If the child is not a leaf, move the last child of the sibling to the child
	if !x.leaf {
		child.children[0] = sibling.children[sibling.n]
	}
	// Move the last key of the sibling to the parent
	x.keys[i-1] = sibling.keys[sibling.n-1]
	// Update the sizes
	child.n++
	sibling.n--
}

// Borrows a key from node x.children[i+1] and includes it in x.children[i]
func (bt *BTree) borrowFromNext(x *BTreeNode, i int) {
	// Get the child and sibling
	child := x.children[i]
	sibling := x.children[i+1]

	// Move the key from the parent to the child
	child.keys[child.n] = x.keys[i]

	// If the child is not a leaf, move the first child of the sibling to the child
	if !child.leaf {
		child.children[child.n+1] = sibling.children[0]
	}

	// Move the first key of the sibling to the parent
	x.keys[i] = sibling.keys[0]

	// Shift all keys in the sibling to the left
	for j := 1; j < sibling.n; j++ {
		sibling.keys[j-1] = sibling.keys[j]
	}

	// If the sibling is not a leaf, shift all children in the sibling to the left
	if !sibling.leaf {
		for j := 1; j <= sibling.n; j++ {
			sibling.children[j-1] = sibling.children[j]
		}
	}

	// Update the sizes
	child.n++
	sibling.n--
}

// Merges x.children[i] and x.children[i+1]
func (bt *BTree) merge(x *BTreeNode, i int) {
	child := x.children[i]
	sibling := x.children[i+1]

	// Ensure child.keys and child.children have enough capacity
	child.keys = append(child.keys, Content{})       // Ensure there's space for the key from x
	child.keys = append(child.keys, sibling.keys...) // Ensure there's space for sibling's keys

	// If not a leaf, check there's space for sibling's children in child and append them
	if !child.leaf {
		child.children = append(child.children, sibling.children...)
	}

	// Move key from x to child to make it the new median
	child.keys[child.n] = x.keys[i]

	// Move keys from sibling to child to keep the order
	for j := 0; j < sibling.n; j++ {
		child.keys[child.n+1+j] = sibling.keys[j]
	}

	// If not a leaf, move children from sibling to child
	if !child.leaf {
		for j := 0; j <= sibling.n; j++ {
			child.children[child.n+1+j] = sibling.children[j]
		}
	}

	// Shift keys in x to the left to fill the gap
	for j := i + 1; j < x.n; j++ {
		x.keys[j-1] = x.keys[j]
	}

	// Shift children to the left to fill the gap
	for j := i + 2; j <= x.n; j++ {
		x.children[j-1] = x.children[j]
	}

	// Update sizes
	child.n += sibling.n + 1
	x.n--
}

// Returns the record with the smallest key
func (bt *BTree) GetSmallest() (Content, error) {
	// Avoid case of a initialized tree with zero nodes
	if bt.root != nil && bt.GetSize() != 0 {
		return getSmallest(bt.root), nil
	} else {
		// Return an error if the tree is empty
		return Content{}, errors.New("Empty BTree")
	}
}

// Main function for GetSmallest
func getSmallest(node *BTreeNode) Content {
	// If it's leaf, dive deeper into the leftmost node
	for !node.leaf {
		node = node.children[0]
	}

	// Return the smallest key in the leaf node
	if node.n > 0 {
		return node.keys[0]
	}
	// Just in case, we add this return
	// Should never happen, this is to avoid returning an empty struct
	return Content{}
}

// Returns the record with the greatest key
func (bt *BTree) GetGreatest() (Content, error) {
	// Case of an empty tree
	if bt.root != nil && bt.GetSize() != 0 {
		return getGreatest(bt.root), nil
	} else {
		// Return an error if the tree is empty
		return Content{}, errors.New("Empty BTree")
	}
}

// Gets the greatest value of the BTree
func getGreatest(node *BTreeNode) Content {
	// If it's leaf, dive deeper into the rightmost node (note the for)
	for !node.leaf {
		node = node.children[node.n]
	}

	// Return the biggest key in the node
	if node.n > 0 {
		return node.keys[node.n-1]
	}
	// Just in case, we add this return (should never happen)
	return Content{}
}

// Returns the size of the tree
func (bt *BTree) GetSize() int {
	return bt.size
}

// Returns the values in the tree in order
func (bt *BTree) ValuesInOrder() []int {
	var output []int
	if bt.GetSize() == 0 {
		return output
	}
	if bt.root != nil {
		output = valuesInOrder(bt.root)
	}
	return output
}

// Main function for ValuesInOrder
func valuesInOrder(node *BTreeNode) []int {
	var result []int

	i := 0
	// Traverse the keys and children
	for i < node.n {
		// If it's not a leaf, we need to go deeper
		if !node.leaf {
			result = append(result, valuesInOrder(node.children[i])...)
		}
		// Append the key and move forward
		result = append(result, node.keys[i].Key)
		i++
	}
	// If it's not a leaf, we need to go deeper
	if !node.leaf {
		result = append(result, valuesInOrder(node.children[i])...)
	}

	return result
}
