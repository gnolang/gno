package bptree

// node is the interface satisfied by both inner and leaf nodes.
type node interface {
	isLeaf() bool
	nodeSize() int // total leaf entries in subtree
	minKey() string
}

//----------------------------------------
// leafNode

type leafNode struct {
	keys   []string
	values []*any // each value is a separate object for lazy loading
}

func newLeafNode(fanout int) *leafNode {
	return &leafNode{
		keys:   make([]string, 0, fanout),
		values: make([]*any, 0, fanout),
	}
}

func (n *leafNode) isLeaf() bool   { return true }
func (n *leafNode) nodeSize() int  { return len(n.keys) }
func (n *leafNode) minKey() string { return n.keys[0] }

// find returns the index where key is or would be inserted,
// and whether an exact match was found.
func (n *leafNode) find(key string) (int, bool) {
	lo, hi := 0, len(n.keys)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if n.keys[mid] < key {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(n.keys) && n.keys[lo] == key {
		return lo, true
	}
	return lo, false
}

// insertAt inserts a key-value pair at the given position.
func (n *leafNode) insertAt(pos int, key string, value *any) {
	n.keys = append(n.keys, "")
	copy(n.keys[pos+1:], n.keys[pos:])
	n.keys[pos] = key

	n.values = append(n.values, nil)
	copy(n.values[pos+1:], n.values[pos:])
	n.values[pos] = value
}

// removeAt removes the entry at the given position and returns the removed key and value.
func (n *leafNode) removeAt(pos int) (string, *any) {
	key := n.keys[pos]
	value := n.values[pos]

	copy(n.keys[pos:], n.keys[pos+1:])
	n.keys[len(n.keys)-1] = ""
	n.keys = n.keys[:len(n.keys)-1]

	copy(n.values[pos:], n.values[pos+1:])
	n.values[len(n.values)-1] = nil
	n.values = n.values[:len(n.values)-1]

	return key, value
}

//----------------------------------------
// innerNode

type innerNode struct {
	keys     []string // separator keys; keys[i] = minKey of children[i+1]
	children []node
	sizes    []int // sizes[i] = total leaf entries in children[i]
}

func newInnerNode(fanout int) *innerNode {
	return &innerNode{
		keys:     make([]string, 0, fanout-1),
		children: make([]node, 0, fanout),
		sizes:    make([]int, 0, fanout),
	}
}

func (n *innerNode) isLeaf() bool { return false }
func (n *innerNode) nodeSize() int {
	sum := 0
	for _, s := range n.sizes {
		sum += s
	}
	return sum
}
func (n *innerNode) minKey() string { return n.children[0].minKey() }

// findChild returns the child index for the given key.
func (n *innerNode) findChild(key string) int {
	// Binary search: find the rightmost i where keys[i] <= key.
	lo, hi := 0, len(n.keys)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if n.keys[mid] <= key {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// insertChildAt inserts a new child with its separator key at the given position.
// The separator key is placed at keys[pos-1] (since keys[i] = minKey of children[i+1]).
func (n *innerNode) insertChildAt(pos int, sep string, child node, sz int) {
	// Insert separator at keys[pos-1].
	keyPos := pos - 1
	n.keys = append(n.keys, "")
	copy(n.keys[keyPos+1:], n.keys[keyPos:])
	n.keys[keyPos] = sep

	// Insert child at children[pos].
	n.children = append(n.children, nil)
	copy(n.children[pos+1:], n.children[pos:])
	n.children[pos] = child

	// Insert size at sizes[pos].
	n.sizes = append(n.sizes, 0)
	copy(n.sizes[pos+1:], n.sizes[pos:])
	n.sizes[pos] = sz
}

// removeChildAt removes the child at pos and its associated separator key.
func (n *innerNode) removeChildAt(pos int) {
	// Determine which separator to remove.
	// keys[i] separates children[i] from children[i+1].
	// Removing children[pos]: if pos > 0, remove keys[pos-1]; else remove keys[0].
	keyPos := pos
	if pos > 0 {
		keyPos = pos - 1
	}
	if len(n.keys) > 0 {
		copy(n.keys[keyPos:], n.keys[keyPos+1:])
		n.keys[len(n.keys)-1] = ""
		n.keys = n.keys[:len(n.keys)-1]
	}

	copy(n.children[pos:], n.children[pos+1:])
	n.children[len(n.children)-1] = nil
	n.children = n.children[:len(n.children)-1]

	copy(n.sizes[pos:], n.sizes[pos+1:])
	n.sizes = n.sizes[:len(n.sizes)-1]
}
