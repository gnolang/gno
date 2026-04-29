package bptree

type ITree interface {
	Size() int
	Has(key string) bool
	Get(key string) (value any, exists bool)
	GetByIndex(index int) (key string, value any)
	Iterate(start, end string, cb IterCbFn) bool
	ReverseIterate(start, end string, cb IterCbFn) bool
	IterateByOffset(offset int, count int, cb IterCbFn) bool
	ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool
	Set(key string, value any) (updated bool)
	Remove(key string) (value any, removed bool)
}

type IterCbFn func(key string, value any) bool

// Verify BPTree implements ITree.
var _ ITree = (*BPTree)(nil)

// The zero value is usable as an empty tree with fanout 32.
type BPTree struct {
	root   node
	size   int
	fanout int
}

// NewBPTreeN creates a new empty B+ tree with the given fanout.
// It panics when fanout is lower than 4.
func NewBPTreeN(fanout int) *BPTree {
	if fanout < 4 {
		panic("bptree: fanout must be >= 4")
	}
	return &BPTree{fanout: fanout}
}

// NewBPTree32 creates a new empty B+ tree with fanout 32.
func NewBPTree32() *BPTree {
	return NewBPTreeN(32)
}

func (t *BPTree) Size() int {
	return t.size
}

func (t *BPTree) Has(key string) bool {
	if t.root == nil {
		return false
	}
	n := t.root
	for !n.isLeaf() {
		inner := n.(*innerNode)
		idx := inner.findChild(key)
		n = inner.children[idx]
	}
	leaf := n.(*leafNode)
	_, found := leaf.find(key)
	return found
}

// Get retrieves the value associated with the given key.
func (t *BPTree) Get(key string) (value any, exists bool) {
	if t.root == nil {
		return nil, false
	}
	n := t.root
	for !n.isLeaf() {
		inner := n.(*innerNode)
		idx := inner.findChild(key)
		n = inner.children[idx]
	}
	leaf := n.(*leafNode)
	pos, found := leaf.find(key)
	if !found {
		return nil, false
	}
	return *leaf.values[pos], true
}

// GetByIndex returns the key-value pair at the given 0-based index.
// Panics if index is out of range.
func (t *BPTree) GetByIndex(index int) (key string, value any) {
	if t.root == nil || index < 0 || index >= t.size {
		panic("GetByIndex asked for invalid index")
	}
	n := t.root
	rem := index
	for !n.isLeaf() {
		inner := n.(*innerNode)
		found := false
		for i, s := range inner.sizes {
			if rem < s {
				n = inner.children[i]
				found = true
				break
			}
			rem -= s
		}
		if !found {
			panic("GetByIndex asked for invalid index")
		}
	}
	leaf := n.(*leafNode)
	return leaf.keys[rem], *leaf.values[rem]
}

//----------------------------------------
// Set

// pathEntry records a step in the root-to-leaf descent.
type pathEntry struct {
	inner    *innerNode
	childIdx int
}

// Set inserts or updates a key-value pair. Returns true if the key already existed.
func (t *BPTree) Set(key string, value any) (updated bool) {
	if t.fanout == 0 {
		t.fanout = 32
	}
	fanout := t.fanout

	vp := &value // wrap value in *any for lazy loading

	// Empty tree: create a single leaf.
	if t.root == nil {
		leaf := newLeafNode(fanout)
		leaf.keys = append(leaf.keys, key)
		leaf.values = append(leaf.values, vp)
		t.root = leaf
		t.size = 1
		return false
	}

	// Descend to the leaf, recording the path.
	var path []pathEntry
	n := t.root
	for !n.isLeaf() {
		inner := n.(*innerNode)
		idx := inner.findChild(key)
		path = append(path, pathEntry{inner, idx})
		n = inner.children[idx]
	}
	leaf := n.(*leafNode)

	// Check if key already exists.
	pos, found := leaf.find(key)
	if found {
		*leaf.values[pos] = value
		return true
	}

	// Insert new key-value.
	leaf.insertAt(pos, key, vp)
	t.size++

	// Increment sizes up the path.
	for i := range path {
		path[i].inner.sizes[path[i].childIdx]++
	}

	// Split if leaf overflows.
	if len(leaf.keys) > fanout {
		t.splitLeaf(leaf, pos, path)
	}

	return false
}

// splitLeaf splits an overflowed leaf node. Uses 90/10 split for append
// patterns (insertPos == fanout) and 50/50 otherwise.
func (t *BPTree) splitLeaf(leaf *leafNode, insertPos int, path []pathEntry) {
	fanout := t.fanout

	// 90/10 split for append pattern: new key ended up at position fanout
	// (greater than all pre-existing keys).
	var mid int
	if insertPos == fanout {
		mid = fanout - 1
	} else {
		mid = (fanout + 1) / 2
	}

	// Create right leaf with entries [mid, fanout+1).
	right := newLeafNode(fanout)
	right.keys = append(right.keys, leaf.keys[mid:]...)
	right.values = append(right.values, leaf.values[mid:]...)

	// Truncate left leaf to [0, mid).
	for i := mid; i < len(leaf.keys); i++ {
		leaf.keys[i] = ""
		leaf.values[i] = nil
	}
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]

	// Promote separator to parent.
	sep := right.keys[0]
	leftSize := len(leaf.keys)
	rightSize := len(right.keys)

	if len(path) == 0 {
		// Root was the leaf; create a new inner root.
		newRoot := newInnerNode(fanout)
		newRoot.keys = append(newRoot.keys, sep)
		newRoot.children = append(newRoot.children, leaf, right)
		newRoot.sizes = append(newRoot.sizes, leftSize, rightSize)
		t.root = newRoot
		return
	}

	// Insert into parent.
	pe := path[len(path)-1]
	parent := pe.inner
	childIdx := pe.childIdx

	parent.sizes[childIdx] = leftSize
	parent.insertChildAt(childIdx+1, sep, right, rightSize)

	if len(parent.children) > fanout {
		t.splitInner(parent, path[:len(path)-1])
	}
}

// splitInner splits an overflowed inner node (always 50/50).
func (t *BPTree) splitInner(inner *innerNode, path []pathEntry) {
	fanout := t.fanout
	mid := (fanout + 1) / 2

	promotedKey := inner.keys[mid-1]

	right := newInnerNode(fanout)
	right.keys = append(right.keys, inner.keys[mid:]...)
	right.children = append(right.children, inner.children[mid:]...)
	right.sizes = append(right.sizes, inner.sizes[mid:]...)

	for i := mid - 1; i < len(inner.keys); i++ {
		inner.keys[i] = ""
	}
	for i := mid; i < len(inner.children); i++ {
		inner.children[i] = nil
	}
	inner.keys = inner.keys[:mid-1]
	inner.children = inner.children[:mid]
	inner.sizes = inner.sizes[:mid]

	if len(path) == 0 {
		newRoot := newInnerNode(fanout)
		newRoot.keys = append(newRoot.keys, promotedKey)
		newRoot.children = append(newRoot.children, inner, right)
		newRoot.sizes = append(newRoot.sizes, inner.nodeSize(), right.nodeSize())
		t.root = newRoot
		return
	}

	pe := path[len(path)-1]
	parent := pe.inner
	childIdx := pe.childIdx

	parent.sizes[childIdx] = inner.nodeSize()
	parent.insertChildAt(childIdx+1, promotedKey, right, right.nodeSize())

	if len(parent.children) > fanout {
		t.splitInner(parent, path[:len(path)-1])
	}
}

//----------------------------------------
// Remove

// Remove deletes a key. Returns the old value and true if the key was found.
func (t *BPTree) Remove(key string) (value any, removed bool) {
	if t.root == nil {
		return nil, false
	}
	fanout := t.fanout

	var path []pathEntry
	n := t.root
	for !n.isLeaf() {
		inner := n.(*innerNode)
		idx := inner.findChild(key)
		path = append(path, pathEntry{inner, idx})
		n = inner.children[idx]
	}
	leaf := n.(*leafNode)

	pos, found := leaf.find(key)
	if !found {
		return nil, false
	}

	var vp *any
	_, vp = leaf.removeAt(pos)
	value = *vp
	t.size--

	for i := range path {
		path[i].inner.sizes[path[i].childIdx]--
	}

	// Handle root leaf.
	if len(path) == 0 {
		if len(leaf.keys) == 0 {
			t.root = nil
		}
		return value, true
	}

	// Check underflow.
	minKeys := fanout / 2
	if len(leaf.keys) >= minKeys {
		if pos == 0 {
			t.updateSeparator(path)
		}
		return value, true
	}

	// If the minimum key was removed, fix ancestor separators before rebalancing.
	if pos == 0 {
		t.updateSeparator(path)
	}

	// Rebalance.
	t.rebalanceLeaf(leaf, path)
	return value, true
}

// updateSeparator fixes the parent's separator key if the child's min key changed
// (e.g., after removing the leftmost entry).
func (t *BPTree) updateSeparator(path []pathEntry) {
	for i := len(path) - 1; i >= 0; i-- {
		pe := path[i]
		if pe.childIdx > 0 {
			child := pe.inner.children[pe.childIdx]
			pe.inner.keys[pe.childIdx-1] = child.minKey()
			return
		}
	}
}

// rebalanceLeaf handles a leaf that has underflowed (< fanout/2 entries).
// Tries redistribute from left, then right, then merges with a sibling.
func (t *BPTree) rebalanceLeaf(leaf *leafNode, path []pathEntry) {
	fanout := t.fanout
	minKeys := fanout / 2
	pe := path[len(path)-1]
	parent := pe.inner
	childIdx := pe.childIdx

	// Try redistribute from left sibling.
	if childIdx > 0 {
		leftSib := parent.children[childIdx-1].(*leafNode)
		if len(leftSib.keys) > minKeys {
			k, v := leftSib.removeAt(len(leftSib.keys) - 1)
			leaf.insertAt(0, k, v)
			parent.keys[childIdx-1] = leaf.keys[0]
			parent.sizes[childIdx-1] = len(leftSib.keys)
			parent.sizes[childIdx] = len(leaf.keys)
			return
		}
	}

	// Try redistribute from right sibling.
	if childIdx < len(parent.children)-1 {
		rightSib := parent.children[childIdx+1].(*leafNode)
		if len(rightSib.keys) > minKeys {
			k, v := rightSib.removeAt(0)
			leaf.insertAt(len(leaf.keys), k, v)
			parent.keys[childIdx] = rightSib.keys[0]
			parent.sizes[childIdx] = len(leaf.keys)
			parent.sizes[childIdx+1] = len(rightSib.keys)
			return
		}
	}

	// Merge.
	if childIdx > 0 {
		leftSib := parent.children[childIdx-1].(*leafNode)
		mergeLeaves(leftSib, leaf, parent, childIdx)
	} else {
		rightSib := parent.children[childIdx+1].(*leafNode)
		mergeLeaves(leaf, rightSib, parent, childIdx+1)
	}

	t.rebalanceInner(path[:len(path)-1])
}

// mergeLeaves merges right leaf into left and removes right from parent.
func mergeLeaves(left, right *leafNode, parent *innerNode, rightIdx int) {
	left.keys = append(left.keys, right.keys...)
	left.values = append(left.values, right.values...)
	parent.removeChildAt(rightIdx)
	parent.sizes[rightIdx-1] = len(left.keys)
}

// rebalanceInner handles an inner node that has underflowed after a child merge.
func (t *BPTree) rebalanceInner(path []pathEntry) {
	if len(path) == 0 {
		root := t.root.(*innerNode)
		if len(root.children) == 1 {
			t.root = root.children[0]
		}
		return
	}

	pe := path[len(path)-1]
	parent := pe.inner
	childIdx := pe.childIdx
	child := parent.children[childIdx].(*innerNode)
	fanout := t.fanout
	minChildren := fanout / 2

	if len(child.children) >= minChildren {
		if childIdx > 0 {
			parent.keys[childIdx-1] = child.minKey()
		}
		return
	}

	// Try redistribute from left sibling.
	if childIdx > 0 {
		leftSib := parent.children[childIdx-1].(*innerNode)
		if len(leftSib.children) > minChildren {
			redistributeInnerLeft(parent, childIdx, leftSib, child)
			return
		}
	}

	// Try redistribute from right sibling.
	if childIdx < len(parent.children)-1 {
		rightSib := parent.children[childIdx+1].(*innerNode)
		if len(rightSib.children) > minChildren {
			redistributeInnerRight(parent, childIdx, child, rightSib)
			return
		}
	}

	// Merge.
	if childIdx > 0 {
		leftSib := parent.children[childIdx-1].(*innerNode)
		mergeInner(leftSib, child, parent, childIdx)
	} else {
		rightSib := parent.children[childIdx+1].(*innerNode)
		mergeInner(child, rightSib, parent, childIdx+1)
	}

	grandPath := path[:len(path)-1]
	if len(grandPath) == 0 {
		root := t.root.(*innerNode)
		if len(root.children) == 1 {
			t.root = root.children[0]
		}
		return
	}

	gpe := grandPath[len(grandPath)-1]
	gparent := gpe.inner
	gchildIdx := gpe.childIdx
	gchild := gparent.children[gchildIdx].(*innerNode)
	minChildren = t.fanout / 2

	if len(gchild.children) < minChildren {
		t.rebalanceInner(grandPath)
	} else if gchildIdx > 0 {
		gparent.keys[gchildIdx-1] = gchild.minKey()
	}
}

// redistributeInnerLeft moves the last child from left sibling to the
// deficient child, pulling down the parent separator and pushing up a new one.
func redistributeInnerLeft(parent *innerNode, childIdx int, leftSib, child *innerNode) {
	sep := parent.keys[childIdx-1]
	child.keys = append(child.keys, "")
	copy(child.keys[1:], child.keys)
	child.keys[0] = sep

	lastChild := leftSib.children[len(leftSib.children)-1]
	lastSize := leftSib.sizes[len(leftSib.sizes)-1]
	child.children = append(child.children, nil)
	copy(child.children[1:], child.children)
	child.children[0] = lastChild
	child.sizes = append(child.sizes, 0)
	copy(child.sizes[1:], child.sizes)
	child.sizes[0] = lastSize

	parent.keys[childIdx-1] = leftSib.keys[len(leftSib.keys)-1]

	leftSib.keys[len(leftSib.keys)-1] = ""
	leftSib.keys = leftSib.keys[:len(leftSib.keys)-1]
	leftSib.children[len(leftSib.children)-1] = nil
	leftSib.children = leftSib.children[:len(leftSib.children)-1]
	leftSib.sizes = leftSib.sizes[:len(leftSib.sizes)-1]

	parent.sizes[childIdx-1] = leftSib.nodeSize()
	parent.sizes[childIdx] = child.nodeSize()
}

// redistributeInnerRight moves the first child from right sibling to the
// deficient child, pulling down the parent separator and pushing up a new one.
func redistributeInnerRight(parent *innerNode, childIdx int, child, rightSib *innerNode) {
	sep := parent.keys[childIdx]
	child.keys = append(child.keys, sep)

	firstChild := rightSib.children[0]
	firstSize := rightSib.sizes[0]
	child.children = append(child.children, firstChild)
	child.sizes = append(child.sizes, firstSize)

	parent.keys[childIdx] = rightSib.keys[0]

	copy(rightSib.keys, rightSib.keys[1:])
	rightSib.keys[len(rightSib.keys)-1] = ""
	rightSib.keys = rightSib.keys[:len(rightSib.keys)-1]
	copy(rightSib.children, rightSib.children[1:])
	rightSib.children[len(rightSib.children)-1] = nil
	rightSib.children = rightSib.children[:len(rightSib.children)-1]
	copy(rightSib.sizes, rightSib.sizes[1:])
	rightSib.sizes = rightSib.sizes[:len(rightSib.sizes)-1]

	parent.sizes[childIdx] = child.nodeSize()
	parent.sizes[childIdx+1] = rightSib.nodeSize()
}

// mergeInner merges right inner node into left, pulling down the parent
// separator, and removes right from parent.
func mergeInner(left, right *innerNode, parent *innerNode, rightIdx int) {
	sep := parent.keys[rightIdx-1]
	left.keys = append(left.keys, sep)
	left.keys = append(left.keys, right.keys...)
	left.children = append(left.children, right.children...)
	left.sizes = append(left.sizes, right.sizes...)
	parent.removeChildAt(rightIdx)
	parent.sizes[rightIdx-1] = left.nodeSize()
}

//----------------------------------------
// Stack-based iteration

// iterStack is a stack of (innerNode, childIdx) for traversal.
// It is ephemeral — built during iteration, discarded after.
type iterStack []pathEntry

// descendLeft descends to the leftmost leaf, pushing inner nodes onto the stack.
func descendLeft(n node, stack *iterStack) *leafNode {
	for !n.isLeaf() {
		inner := n.(*innerNode)
		*stack = append(*stack, pathEntry{inner, 0})
		n = inner.children[0]
	}
	return n.(*leafNode)
}

// descendRight descends to the rightmost leaf, pushing inner nodes onto the stack.
func descendRight(n node, stack *iterStack) *leafNode {
	for !n.isLeaf() {
		inner := n.(*innerNode)
		last := len(inner.children) - 1
		*stack = append(*stack, pathEntry{inner, last})
		n = inner.children[last]
	}
	return n.(*leafNode)
}

// advanceLeaf moves to the next leaf in ascending order via the stack.
// Returns nil if there is no next leaf.
func advanceLeaf(stack *iterStack) *leafNode {
	for len(*stack) > 0 {
		top := &(*stack)[len(*stack)-1]
		top.childIdx++
		if top.childIdx < len(top.inner.children) {
			n := top.inner.children[top.childIdx]
			return descendLeft(n, stack)
		}
		*stack = (*stack)[:len(*stack)-1]
	}
	return nil
}

// retreatLeaf moves to the previous leaf in descending order via the stack.
// Returns nil if there is no previous leaf.
func retreatLeaf(stack *iterStack) *leafNode {
	for len(*stack) > 0 {
		top := &(*stack)[len(*stack)-1]
		top.childIdx--
		if top.childIdx >= 0 {
			n := top.inner.children[top.childIdx]
			return descendRight(n, stack)
		}
		*stack = (*stack)[:len(*stack)-1]
	}
	return nil
}

//----------------------------------------
// Iterate / ReverseIterate

// Iterate calls cb for each key-value pair in [start, end) ascending order.
// Empty start/end means no bound. Returns true if stopped early by cb.
// The tree must not be modified during iteration (no Set or Remove from the callback).
func (t *BPTree) Iterate(start, end string, cb IterCbFn) bool {
	if t.root == nil {
		return false
	}

	var stack iterStack
	var leaf *leafNode
	var pos int

	if start == "" {
		leaf = descendLeft(t.root, &stack)
		pos = 0
	} else {
		leaf, pos, stack = t.descendToGE(start)
		if leaf == nil {
			return false
		}
	}

	for leaf != nil {
		for pos < len(leaf.keys) {
			k := leaf.keys[pos]
			if end != "" && k >= end {
				return false
			}
			if cb(k, *leaf.values[pos]) {
				return true
			}
			pos++
		}
		leaf = advanceLeaf(&stack)
		pos = 0
	}
	return false
}

// ReverseIterate calls cb for each key-value pair in [start, end] descending order.
// Empty start/end means no bound. Returns true if stopped early by cb.
// The tree must not be modified during iteration (no Set or Remove from the callback).
func (t *BPTree) ReverseIterate(start, end string, cb IterCbFn) bool {
	if t.root == nil {
		return false
	}

	var stack iterStack
	var leaf *leafNode
	var pos int

	if end == "" {
		leaf = descendRight(t.root, &stack)
		pos = len(leaf.keys) - 1
	} else {
		leaf, pos, stack = t.descendToLE(end)
		if leaf == nil {
			return false
		}
	}

	for leaf != nil {
		for pos >= 0 {
			k := leaf.keys[pos]
			if start != "" && k < start {
				return false
			}
			if cb(k, *leaf.values[pos]) {
				return true
			}
			pos--
		}
		leaf = retreatLeaf(&stack)
		if leaf != nil {
			pos = len(leaf.keys) - 1
		}
	}
	return false
}

// IterateByOffset calls cb for count entries starting at the offset-th entry
// in ascending order. Returns true if stopped early by cb.
// The tree must not be modified during iteration (no Set or Remove from the callback).
func (t *BPTree) IterateByOffset(offset int, count int, cb IterCbFn) bool {
	if t.root == nil || offset >= t.size || count <= 0 {
		return false
	}
	if offset < 0 {
		offset = 0
	}

	leaf, pos, stack := t.descendToOffset(offset)

	visited := 0
	for leaf != nil && visited < count {
		for pos < len(leaf.keys) && visited < count {
			if cb(leaf.keys[pos], *leaf.values[pos]) {
				return true
			}
			visited++
			pos++
		}
		leaf = advanceLeaf(&stack)
		pos = 0
	}
	return false
}

// ReverseIterateByOffset calls cb for count entries starting at the offset-th
// entry from the end, in descending order. Returns true if stopped early by cb.
// The tree must not be modified during iteration (no Set or Remove from the callback).
func (t *BPTree) ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool {
	if t.root == nil || offset >= t.size || count <= 0 {
		return false
	}
	if offset < 0 {
		offset = 0
	}

	ascIdx := t.size - 1 - offset
	leaf, pos, stack := t.descendToOffset(ascIdx)

	visited := 0
	for leaf != nil && visited < count {
		for pos >= 0 && visited < count {
			if cb(leaf.keys[pos], *leaf.values[pos]) {
				return true
			}
			visited++
			pos--
		}
		leaf = retreatLeaf(&stack)
		if leaf != nil {
			pos = len(leaf.keys) - 1
		}
	}
	return false
}

//----------------------------------------
// Descent helpers for iteration

// descendToGE descends to the first key >= key, returning the leaf, position, and stack.
func (t *BPTree) descendToGE(key string) (*leafNode, int, iterStack) {
	var stack iterStack
	n := t.root
	for !n.isLeaf() {
		inner := n.(*innerNode)
		idx := inner.findChild(key)
		stack = append(stack, pathEntry{inner, idx})
		n = inner.children[idx]
	}
	leaf := n.(*leafNode)
	pos, _ := leaf.find(key)
	if pos >= len(leaf.keys) {
		next := advanceLeaf(&stack)
		if next == nil {
			return nil, 0, nil
		}
		return next, 0, stack
	}
	return leaf, pos, stack
}

// descendToLE descends to the last key <= key, returning the leaf, position, and stack.
func (t *BPTree) descendToLE(key string) (*leafNode, int, iterStack) {
	var stack iterStack
	n := t.root
	for !n.isLeaf() {
		inner := n.(*innerNode)
		idx := inner.findChild(key)
		stack = append(stack, pathEntry{inner, idx})
		n = inner.children[idx]
	}
	leaf := n.(*leafNode)
	pos, found := leaf.find(key)
	if found {
		return leaf, pos, stack
	}
	if pos > 0 {
		return leaf, pos - 1, stack
	}
	prev := retreatLeaf(&stack)
	if prev == nil {
		return nil, 0, nil
	}
	return prev, len(prev.keys) - 1, stack
}

// descendToOffset descends to the offset-th entry using sizes[],
// returning the leaf, position within leaf, and stack.
func (t *BPTree) descendToOffset(offset int) (*leafNode, int, iterStack) {
	var stack iterStack
	n := t.root
	rem := offset
	for !n.isLeaf() {
		inner := n.(*innerNode)
		found := false
		for i, s := range inner.sizes {
			if rem < s {
				stack = append(stack, pathEntry{inner, i})
				n = inner.children[i]
				found = true
				break
			}
			rem -= s
		}
		if !found {
			panic("descendToOffset: offset out of range")
		}
	}
	return n.(*leafNode), rem, stack
}
