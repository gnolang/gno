package bptree

import "bytes"

// searchLeaf finds the position of key in a leaf node's sorted keys.
// Returns (index, found). If found, keys[index] == key.
// If not found, index is where the key would be inserted.
func searchLeaf(n *LeafNode, key []byte) (int, bool) {
	lo, hi := 0, int(n.numKeys)
	for lo < hi {
		mid := lo + (hi-lo)/2
		cmp := bytes.Compare(n.keys[mid], key)
		if cmp == 0 {
			return mid, true
		}
		if cmp < 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo, false
}

// searchInner finds which child to descend into for the given key.
// Returns the child index (0..numKeys). The invariant is:
//
//	keys[i-1] <= key < keys[i]
//
// meaning child[i] covers keys in [keys[i-1], keys[i]).
func searchInner(n *InnerNode, key []byte) int {
	lo, hi := 0, int(n.numKeys)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if bytes.Compare(n.keys[mid], key) <= 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}
