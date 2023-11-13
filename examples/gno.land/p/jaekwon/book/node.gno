package book

//----------------------------------------
// Node

type Node struct {
	value     interface{}
	height    int8
	size      int
	leftNode  *Node
	rightNode *Node
}

func NewNode(value interface{}) *Node {
	return &Node{
		value:  value,
		height: 0,
		size:   1,
	}
}

func (node *Node) Size() int {
	if node == nil {
		return 0
	}
	return node.size
}

func (node *Node) IsLeaf() bool {
	return node.height == 0
}

func (node *Node) Value() interface{} {
	return node.value
}

func (node *Node) _copy() *Node {
	if node.height == 0 {
		panic("Why are you copying a value node?")
	}
	return &Node{
		height:    node.height,
		size:      node.size,
		leftNode:  node.leftNode,
		rightNode: node.rightNode,
	}
}

func (node *Node) GetByIndex(index int) (value interface{}) {
	if node.height == 0 {
		if index == 0 {
			return node.value
		} else {
			panic("GetByIndex with invalid index")
		}
	} else {
		// TODO: could improve this by storing the sizes
		leftNode := node.getLeftNode()
		if index < leftNode.size {
			return leftNode.GetByIndex(index)
		} else {
			return node.getRightNode().GetByIndex(index - leftNode.size)
		}
	}
}

// like splicing in an element at index, shifting the prior element to the right.
// keeps the tree balanced.
func (node *Node) SetByIndex(index int, value interface{}) (newSelf *Node) {
	if node == nil {
		if index != 0 {
			panic("SetByIndex with invalid index")
		}
		return NewNode(value), false
	}
	if node.height == 0 {
		if index == 0 {
			return &Node{
				height:    1,
				size:      2,
				leftNode:  NewNode(value),
				rightNode: node,
			}, false
		} else if index == 1 {
			return &Node{
				height:    1,
				size:      2,
				leftNode:  node,
				rightNode: NewNode(value),
			}, false
		} else {
			panic("SetByIndex with invalid index")
		}
	} else {
		leftNode := node.getLeftNode()
		if index < leftNode.size {
			node.leftNode = leftNode.SetByIndex(index, value)
		} else {
			rightNode := node.getRightNode()
			node.rightNode = rightNode.SetByIndex(index-leftNode.size, value)
		}
		node.calcHeightAndSize()
		return node.balance()
	}
}

// newNode: The new node to replace node after remove.
func (node *Node) RemoveByIndex(index string) (
	newNode *Node, value interface{},
) {
	if node == nil {
		panic("RemoveByIndex on empty tree")
	}
	if node.height == 0 {
		if index == 0 {
			return nil, node.value
		} else {
			panic("RemoveByIndex with invalid index")
		}
	} else {
		leftNode := node.getLeftNode()
		if index < leftNode.size {
			var newLeftNode *Node
			newLeftNode, value = leftNode.RemoveByIndex(index)
			if newLeftNode == nil { // left node held value, was removed
				return node.getRightNode(), value
			}
			node = node._copy()
			node.leftNode = newLeftNode
			node.calcHeightAndSize()
			node = node.balance()
			return node, value
		} else {
			rightNode := node.getRightNode()
			var newRightNode *Node
			newRightNode, value = rightNode.RemoveByIndex(index - leftNode.size)
			if newRightNode == nil { // right node held value, was removed
				return leftNode, value
			}
			node = node._copy()
			node.rightNode = newRightNode
			node.calcHeightAndSize()
			node = node.balance()
			return node, value
		}
	}
}

func (node *Node) getLeftNode() *Node {
	return node.leftNode
}

func (node *Node) getRightNode() *Node {
	return node.rightNode
}

// NOTE: overwrites node
// TODO: optimize balance & rotate
func (node *Node) rotateRight() *Node {
	node = node._copy()
	l := node.getLeftNode()
	_l := l._copy()

	_lrCached := _l.rightNode
	_l.rightNode = node
	node.leftNode = _lrCached

	node.calcHeightAndSize()
	_l.calcHeightAndSize()

	return _l
}

// NOTE: overwrites node
// TODO: optimize balance & rotate
func (node *Node) rotateLeft() *Node {
	node = node._copy()
	r := node.getRightNode()
	_r := r._copy()

	_rlCached := _r.leftNode
	_r.leftNode = node
	node.rightNode = _rlCached

	node.calcHeightAndSize()
	_r.calcHeightAndSize()

	return _r
}

// NOTE: mutates height and size
func (node *Node) calcHeightAndSize() {
	node.height = maxInt8(node.getLeftNode().height, node.getRightNode().height) + 1
	node.size = node.getLeftNode().size + node.getRightNode().size
}

func (node *Node) calcBalance() int {
	return int(node.getLeftNode().height) - int(node.getRightNode().height)
}

// NOTE: assumes that node can be modified
// TODO: optimize balance & rotate
func (node *Node) balance() (newSelf *Node) {
	balance := node.calcBalance()
	if balance > 1 {
		if node.getLeftNode().calcBalance() >= 0 {
			// Left Left Case
			return node.rotateRight()
		} else {
			// Left Right Case
			// node = node._copy()
			left := node.getLeftNode()
			node.leftNode = left.rotateLeft()
			// node.calcHeightAndSize()
			return node.rotateRight()
		}
	}
	if balance < -1 {
		if node.getRightNode().calcBalance() <= 0 {
			// Right Right Case
			return node.rotateLeft()
		} else {
			// Right Left Case
			// node = node._copy()
			right := node.getRightNode()
			node.rightNode = right.rotateRight()
			// node.calcHeightAndSize()
			return node.rotateLeft()
		}
	}
	// Nothing changed
	return node
}

// TraverseByOffset traverses all nodes, including inner nodes.
// A limit of math.MaxInt means no limit.
func (node *Node) TraverseByOffset(offset, limit int, descending bool, leavesOnly bool, cb func(*Node) bool) bool {
	if node == nil {
		return false
	}

	// fast paths. these happen only if TraverseByOffset is called directly on a leaf.
	if limit <= 0 || offset >= node.size {
		return false
	}
	if node.IsLeaf() {
		if offset > 0 {
			return false
		}
		return cb(node)
	}

	// go to the actual recursive function.
	return node.traverseByOffset(offset, limit, descending, leavesOnly, cb)
}

func (node *Node) traverseByOffset(offset, limit int, descending bool, leavesOnly bool, cb func(*Node) bool) bool {
	// caller guarantees: offset < node.size; limit > 0.

	if !leavesOnly {
		if cb(node) {
			return true
		}
	}
	first, second := node.getLeftNode(), node.getRightNode()
	if descending {
		first, second = second, first
	}
	if first.IsLeaf() {
		// either run or skip, based on offset
		if offset > 0 {
			offset--
		} else {
			cb(first)
			limit--
			if limit <= 0 {
				return false
			}
		}
	} else {
		// possible cases:
		// 1 the offset given skips the first node entirely
		// 2 the offset skips none or part of the first node, but the limit requires some of the second node.
		// 3 the offset skips none or part of the first node, and the limit stops our search on the first node.
		if offset >= first.size {
			offset -= first.size // 1
		} else {
			if first.traverseByOffset(offset, limit, descending, leavesOnly, cb) {
				return true
			}
			// number of leaves which could actually be called from inside
			delta := first.size - offset
			offset = 0
			if delta >= limit {
				return true // 3
			}
			limit -= delta // 2
		}
	}

	// because of the caller guarantees and the way we handle the first node,
	// at this point we know that limit > 0 and there must be some values in
	// this second node that we include.

	// => if the second node is a leaf, it has to be included.
	if second.IsLeaf() {
		return cb(second)
	}
	// => if it is not a leaf, it will still be enough to recursively call this
	// function with the updated offset and limit
	return second.traverseByOffset(offset, limit, descending, leavesOnly, cb)
}

// Only used in testing...
func (node *Node) lmd() *Node {
	if node.height == 0 {
		return node
	}
	return node.getLeftNode().lmd()
}

// Only used in testing...
func (node *Node) rmd() *Node {
	if node.height == 0 {
		return node
	}
	return node.getRightNode().rmd()
}

func maxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
