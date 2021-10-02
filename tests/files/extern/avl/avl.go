package avl

// Node

type Node struct {
	key       string
	value     interface{}
	height    int8
	size      int
	leftNode  *Node
	rightNode *Node
}

func NewNode(key string, value interface{}) *Node {
	return &Node{
		key:    key,
		value:  value,
		height: 0,
		size:   1,
	}
}

func (node *Node) Size() int {
	return node.size
}

func (node *Node) _copy() *Node {
	if node.height == 0 {
		panic("Why are you copying a value node?")
	}
	return &Node{
		key:       node.key,
		height:    node.height,
		size:      node.size,
		leftNode:  node.leftNode,
		rightNode: node.rightNode,
	}
}

func (node *Node) Has(key string) (has bool) {
	if node.key == key {
		return true
	}
	if node.height == 0 {
		return false
	} else {
		if key < node.key {
			return node.getLeftNode().Has(key)
		} else {
			return node.getRightNode().Has(key)
		}
	}
}

func (node *Node) Get(key string) (index int, value interface{}, exists bool) {
	if node.height == 0 {
		if node.key == key {
			return 0, node.value, true
		} else if node.key < key {
			return 1, nil, false
		} else {
			return 0, nil, false
		}
	} else {
		if key < node.key {
			return node.getLeftNode().Get(key)
		} else {
			rightNode := node.getRightNode()
			index, value, exists = rightNode.Get(key)
			index += node.size - rightNode.size
			return index, value, exists
		}
	}
}

func (node *Node) GetByIndex(index int) (key string, value interface{}) {
	if node.height == 0 {
		if index == 0 {
			return node.key, node.value
		} else {
			panic("GetByIndex asked for invalid index")
			return "", nil
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

func (node *Node) Set(key string, value interface{}) (newSelf *Node, updated bool) {
	if node.height == 0 {
		if key < node.key {
			return &Node{
				key:       node.key,
				height:    1,
				size:      2,
				leftNode:  NewNode(key, value),
				rightNode: node,
			}, false
		} else if key == node.key {
			return NewNode(key, value), true
		} else {
			return &Node{
				key:       key,
				height:    1,
				size:      2,
				leftNode:  node,
				rightNode: NewNode(key, value),
			}, false
		}
	} else {
		node = node._copy()
		if key < node.key {
			node.leftNode, updated = node.getLeftNode().Set(key, value)
		} else {
			node.rightNode, updated = node.getRightNode().Set(key, value)
		}
		if updated {
			return node, updated
		} else {
			node.calcHeightAndSize()
			return node.balance(), updated
		}
	}
}

// newNode: The new node to replace node after remove.
// newKey: new leftmost leaf key for tree after successfully removing 'key' if changed.
// value: removed value.
func (node *Node) Remove(key string) (
	newNode *Node, newKey string, value interface{}, removed bool) {
	if node.height == 0 {
		if key == node.key {
			return nil, "", node.value, true
		} else {
			return node, "", nil, false
		}
	} else {
		if key < node.key {
			var newLeftNode *Node
			newLeftNode, newKey, value, removed = node.getLeftNode().Remove(key)
			if !removed {
				return node, "", value, false
			} else if newLeftNode == nil { // left node held value, was removed
				return node.rightNode, node.key, value, true
			}
			node = node._copy()
			node.leftNode = newLeftNode
			node.calcHeightAndSize()
			node = node.balance()
			return node, newKey, value, true
		} else {
			var newRightNode *Node
			newRightNode, newKey, value, removed = node.getRightNode().Remove(key)
			if !removed {
				return node, "", value, false
			} else if newRightNode == nil { // right node held value, was removed
				return node.leftNode, "", value, true
			}
			node = node._copy()
			node.rightNode = newRightNode
			if newKey != "" {
				node.key = newKey
			}
			node.calcHeightAndSize()
			node = node.balance()
			return node, "", value, true
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
			//node.calcHeightAndSize()
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
			//node.calcHeightAndSize()
			return node.rotateLeft()
		}
	}
	// Nothing changed
	return node
}

// traverse is a wrapper over traverseInRange when we want the whole tree
func (node *Node) Traverse(ascending bool, cb func(*Node) bool) bool {
	return node.TraverseInRange("", "", ascending, cb)
}

func (node *Node) TraverseInRange(start, end string, ascending bool, cb func(*Node) bool) bool {
	afterStart := (start == "" || start <= node.key)
	beforeEnd := (end == "" || node.key <= end)

	stop := false
	if afterStart && beforeEnd {
		// IterateRange ignores this if not leaf
		stop = cb(node)
	}
	if stop {
		return stop
	}

	if node.height > 0 {
		if ascending {
			// check lower nodes, then higher
			if afterStart {
				stop = node.getLeftNode().TraverseInRange(start, end, ascending, cb)
			}
			if stop {
				return stop
			}
			if beforeEnd {
				stop = node.getRightNode().TraverseInRange(start, end, ascending, cb)
			}
		} else {
			// check the higher nodes first
			if beforeEnd {
				stop = node.getRightNode().TraverseInRange(start, end, ascending, cb)
			}
			if stop {
				return stop
			}
			if afterStart {
				stop = node.getLeftNode().TraverseInRange(start, end, ascending, cb)
			}
		}
	}

	return stop
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
