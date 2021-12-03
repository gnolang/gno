package avl

// Tree

type Tree struct {
	key       string
	value     interface{}
	height    int8
	size      int
	leftTree  *Tree
	rightTree *Tree
}

func NewTree(key string, value interface{}) *Tree {
	return &Tree{
		key:    key,
		value:  value,
		height: 0,
		size:   1,
	}
}

func (tree *Tree) Size() int {
	if tree == nil {
		return 0
	}
	return tree.size
}

func (tree *Tree) Value() interface{} {
	return tree.value
}

func (tree *Tree) _copy() *Tree {
	if tree.height == 0 {
		panic("Why are you copying a value tree?")
	}
	return &Tree{
		key:       tree.key,
		height:    tree.height,
		size:      tree.size,
		leftTree:  tree.leftTree,
		rightTree: tree.rightTree,
	}
}

func (tree *Tree) Has(key string) (has bool) {
	if tree == nil {
		return false
	}
	if tree.key == key {
		return true
	}
	if tree.height == 0 {
		return false
	} else {
		if key < tree.key {
			return tree.getLeftTree().Has(key)
		} else {
			return tree.getRightTree().Has(key)
		}
	}
}

func (tree *Tree) Get(key string) (index int, value interface{}, exists bool) {
	if tree == nil {
		return 0, nil, false
	}
	if tree.height == 0 {
		if tree.key == key {
			return 0, tree.value, true
		} else if tree.key < key {
			return 1, nil, false
		} else {
			return 0, nil, false
		}
	} else {
		if key < tree.key {
			return tree.getLeftTree().Get(key)
		} else {
			rightTree := tree.getRightTree()
			index, value, exists = rightTree.Get(key)
			index += tree.size - rightTree.size
			return index, value, exists
		}
	}
}

func (tree *Tree) GetByIndex(index int) (key string, value interface{}) {
	if tree.height == 0 {
		if index == 0 {
			return tree.key, tree.value
		} else {
			panic("GetByIndex asked for invalid index")
			return "", nil
		}
	} else {
		// TODO: could improve this by storing the sizes
		leftTree := tree.getLeftTree()
		if index < leftTree.size {
			return leftTree.GetByIndex(index)
		} else {
			return tree.getRightTree().GetByIndex(index - leftTree.size)
		}
	}
}

// XXX consider a better way to do this... perhaps split Tree from Node.
func (tree *Tree) Set(key string, value interface{}) (newSelf *Tree, updated bool) {
	if tree == nil {
		return NewTree(key, value), false
	}
	if tree.height == 0 {
		if key < tree.key {
			return &Tree{
				key:       tree.key,
				height:    1,
				size:      2,
				leftTree:  NewTree(key, value),
				rightTree: tree,
			}, false
		} else if key == tree.key {
			return NewTree(key, value), true
		} else {
			return &Tree{
				key:       key,
				height:    1,
				size:      2,
				leftTree:  tree,
				rightTree: NewTree(key, value),
			}, false
		}
	} else {
		tree = tree._copy()
		if key < tree.key {
			tree.leftTree, updated = tree.getLeftTree().Set(key, value)
		} else {
			tree.rightTree, updated = tree.getRightTree().Set(key, value)
		}
		if updated {
			return tree, updated
		} else {
			tree.calcHeightAndSize()
			return tree.balance(), updated
		}
	}
}

// newTree: The new tree to replace tree after remove.
// newKey: new leftmost leaf key for tree after successfully removing 'key' if changed.
// value: removed value.
func (tree *Tree) Remove(key string) (
	newTree *Tree, newKey string, value interface{}, removed bool) {
	if tree == nil {
		return nil, "", nil, false
	}
	if tree.height == 0 {
		if key == tree.key {
			return nil, "", tree.value, true
		} else {
			return tree, "", nil, false
		}
	} else {
		if key < tree.key {
			var newLeftTree *Tree
			newLeftTree, newKey, value, removed = tree.getLeftTree().Remove(key)
			if !removed {
				return tree, "", value, false
			} else if newLeftTree == nil { // left tree held value, was removed
				return tree.rightTree, tree.key, value, true
			}
			tree = tree._copy()
			tree.leftTree = newLeftTree
			tree.calcHeightAndSize()
			tree = tree.balance()
			return tree, newKey, value, true
		} else {
			var newRightTree *Tree
			newRightTree, newKey, value, removed = tree.getRightTree().Remove(key)
			if !removed {
				return tree, "", value, false
			} else if newRightTree == nil { // right tree held value, was removed
				return tree.leftTree, "", value, true
			}
			tree = tree._copy()
			tree.rightTree = newRightTree
			if newKey != "" {
				tree.key = newKey
			}
			tree.calcHeightAndSize()
			tree = tree.balance()
			return tree, "", value, true
		}
	}
}

func (tree *Tree) getLeftTree() *Tree {
	return tree.leftTree
}

func (tree *Tree) getRightTree() *Tree {
	return tree.rightTree
}

// NOTE: overwrites tree
// TODO: optimize balance & rotate
func (tree *Tree) rotateRight() *Tree {
	tree = tree._copy()
	l := tree.getLeftTree()
	_l := l._copy()

	_lrCached := _l.rightTree
	_l.rightTree = tree
	tree.leftTree = _lrCached

	tree.calcHeightAndSize()
	_l.calcHeightAndSize()

	return _l
}

// NOTE: overwrites tree
// TODO: optimize balance & rotate
func (tree *Tree) rotateLeft() *Tree {
	tree = tree._copy()
	r := tree.getRightTree()
	_r := r._copy()

	_rlCached := _r.leftTree
	_r.leftTree = tree
	tree.rightTree = _rlCached

	tree.calcHeightAndSize()
	_r.calcHeightAndSize()

	return _r
}

// NOTE: mutates height and size
func (tree *Tree) calcHeightAndSize() {
	tree.height = maxInt8(tree.getLeftTree().height, tree.getRightTree().height) + 1
	tree.size = tree.getLeftTree().size + tree.getRightTree().size
}

func (tree *Tree) calcBalance() int {
	return int(tree.getLeftTree().height) - int(tree.getRightTree().height)
}

// NOTE: assumes that tree can be modified
// TODO: optimize balance & rotate
func (tree *Tree) balance() (newSelf *Tree) {
	balance := tree.calcBalance()
	if balance > 1 {
		if tree.getLeftTree().calcBalance() >= 0 {
			// Left Left Case
			return tree.rotateRight()
		} else {
			// Left Right Case
			// tree = tree._copy()
			left := tree.getLeftTree()
			tree.leftTree = left.rotateLeft()
			//tree.calcHeightAndSize()
			return tree.rotateRight()
		}
	}
	if balance < -1 {
		if tree.getRightTree().calcBalance() <= 0 {
			// Right Right Case
			return tree.rotateLeft()
		} else {
			// Right Left Case
			// tree = tree._copy()
			right := tree.getRightTree()
			tree.rightTree = right.rotateRight()
			//tree.calcHeightAndSize()
			return tree.rotateLeft()
		}
	}
	// Nothing changed
	return tree
}

// traverse is a wrapper over traverseInRange when we want the whole tree
func (tree *Tree) Traverse(ascending bool, cb func(*Tree) bool) bool {
	return tree.TraverseInRange("", "", ascending, cb)
}

func (tree *Tree) TraverseInRange(start, end string, ascending bool, cb func(*Tree) bool) bool {
	if tree == nil {
		return false
	}
	afterStart := (start == "" || start <= tree.key)
	beforeEnd := (end == "" || tree.key <= end)

	stop := false
	if afterStart && beforeEnd {
		// IterateRange ignores this if not leaf
		stop = cb(tree)
	}
	if stop {
		return stop
	}

	if tree.height > 0 {
		if ascending {
			// check lower trees, then higher
			if afterStart {
				stop = tree.getLeftTree().TraverseInRange(start, end, ascending, cb)
			}
			if stop {
				return stop
			}
			if beforeEnd {
				stop = tree.getRightTree().TraverseInRange(start, end, ascending, cb)
			}
		} else {
			// check the higher trees first
			if beforeEnd {
				stop = tree.getRightTree().TraverseInRange(start, end, ascending, cb)
			}
			if stop {
				return stop
			}
			if afterStart {
				stop = tree.getLeftTree().TraverseInRange(start, end, ascending, cb)
			}
		}
	}

	return stop
}

// Only used in testing...
func (tree *Tree) lmd() *Tree {
	if tree.height == 0 {
		return tree
	}
	return tree.getLeftTree().lmd()
}

// Only used in testing...
func (tree *Tree) rmd() *Tree {
	if tree.height == 0 {
		return tree
	}
	return tree.getRightTree().rmd()
}

func maxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
