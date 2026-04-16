package bptree

// MiniMerkle is a binary merkle tree over B slots, stored as a heap-style
// array of size 2*B. Index 1 is the root. Indices B..2B-1 are the leaf-level
// slots. Index 0 is unused.
//
// The tree uses the sentinel short-circuit rule: if both children are the
// sentinel, the parent is the sentinel (not SHA256(0x01 || sentinel || sentinel)).
// This ensures ICS23 EmptyChild compatibility at all depths.
type MiniMerkle struct {
	tree [2 * B]Hash
}

// Root returns the mini merkle root hash (index 1).
func (m *MiniMerkle) Root() Hash {
	return m.tree[1]
}

// SetSlot sets the hash at leaf-level slot index (0..B-1) and recomputes
// the path from that slot to the root. Cost: log₂(B) = 5 SHA256 calls.
func (m *MiniMerkle) SetSlot(index int, h Hash) {
	pos := B + index // leaf position in heap array
	m.tree[pos] = h
	// Walk up to root, recomputing parents
	for pos > 1 {
		pos /= 2
		left := m.tree[pos*2]
		right := m.tree[pos*2+1]
		m.tree[pos] = HashInner(left, right)
	}
}

// GetSlot returns the hash at leaf-level slot index (0..B-1).
func (m *MiniMerkle) GetSlot(index int) Hash {
	return m.tree[B+index]
}

// Build recomputes the entire mini merkle tree from the leaf-level slots.
// Cost: B-1 = 31 SHA256 calls for B=32.
func (m *MiniMerkle) Build() {
	for i := B - 1; i >= 1; i-- {
		left := m.tree[i*2]
		right := m.tree[i*2+1]
		m.tree[i] = HashInner(left, right)
	}
}

// emptyMiniMerkle is a MiniMerkle with every slot pre-filled with the
// sentinel hash. Populated once at package init (sentinelHash itself is
// set there) so Clear() can reset a tree via a single struct copy
// (compiles to memcpy) rather than a 64-iteration loop of 32-byte
// writes. See Finding #21.
var emptyMiniMerkle MiniMerkle

// Clear sets all slots to the sentinel hash.
func (m *MiniMerkle) Clear() {
	*m = emptyMiniMerkle
}

// SiblingPath returns the MiniMerkleDepth sibling hashes needed to prove
// that slot[index] is part of the mini merkle root. The path goes from
// the leaf level toward the root. Each entry is the sibling's hash at
// that level. Also returns the position indices (0=left child,
// 1=right child) indicating which side the proven slot is on at each
// level.
//
// Returns fixed-size arrays by value so the results stay on the caller's
// stack; the previous implementation made two parallel slices per call
// via `make(..., 0, 5)`, allocating on every proof step. See Finding #21.
func (m *MiniMerkle) SiblingPath(index int) (siblings [MiniMerkleDepth]Hash, positions [MiniMerkleDepth]int) {
	pos := B + index
	for i := 0; pos > 1; i++ {
		if pos%2 == 0 {
			// pos is left child, sibling is right
			siblings[i] = m.tree[pos+1]
			positions[i] = 0 // proven node is left child
		} else {
			// pos is right child, sibling is left
			siblings[i] = m.tree[pos-1]
			positions[i] = 1 // proven node is right child
		}
		pos /= 2
	}
	return
}

// NewMiniMerkle creates a MiniMerkle with all slots set to the sentinel.
func NewMiniMerkle() MiniMerkle {
	var m MiniMerkle
	m.Clear()
	return m
}
