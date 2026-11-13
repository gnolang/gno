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

// slots returns the leaf-level slot hashes as a mutable view.
//
// The slot hashes are a cache of the per-entry hashes (HashLeafSlot for a
// leaf, childHashes for an inner), and — this is the part the write paths
// exploit — a leaf slot hash is a function of (key, valueHash) only, NOT of
// the slot's index. So an insert or delete that shifts entries within a node
// can shift their slot hashes alongside them instead of recomputing one
// SHA256 per occupied slot.
func (m *MiniMerkle) slots() []Hash {
	return m.tree[B : 2*B]
}

// setSlots overwrites slots [lo, hi] from src (which must be at least
// hi-lo+1 long) and refolds just their ancestors. The whole-node variant is
// Build; this one exists for the write paths that know which slots moved.
func (m *MiniMerkle) setSlots(lo, hi int, src []Hash) {
	copy(m.tree[B+lo:B+hi+1], src[:hi-lo+1])
	m.buildRange(lo, hi)
}

// insertSlot opens a hole at slot pos among count occupied slots, fills it
// with h, and refolds. It is the mini-merkle half of an insert's key shift:
// a slot hash is a function of the entry and NOT of its index, so the
// displaced entries keep theirs and only the new one needs hashing.
func (m *MiniMerkle) insertSlot(pos, count int, h Hash) {
	if count > pos {
		copy(m.tree[B+pos+1:B+count+1], m.tree[B+pos:B+count])
	}
	m.tree[B+pos] = h
	m.buildRange(pos, count)
}

// removeSlot drops slot pos among count occupied slots, sentinels the vacated
// tail, and refolds. The mini-merkle half of a delete's key shift.
func (m *MiniMerkle) removeSlot(pos, count int) {
	if count-1 > pos {
		copy(m.tree[B+pos:B+count-1], m.tree[B+pos+1:B+count])
	}
	m.tree[B+count-1] = sentinelHash
	m.buildRange(pos, count-1)
}

// buildRange recomputes only the internal nodes that cover leaf-level slots
// [lo, hi]. Cost is (hi-lo) + log₂(B) hashes rather than Build's B-1, which
// matters because the common mutations touch a suffix of one node's slots:
// an append-shaped insert (sorted keys, so pos == numKeys) comes out at 5
// hashes instead of 31.
//
// Private, and reached only through Build/setSlots/insertSlot/removeSlot:
// a slot write that is not followed by the matching refold leaves the root
// stale, so no caller gets to sequence the two steps itself.
func (m *MiniMerkle) buildRange(lo, hi int) {
	l, h := B+lo, B+hi
	for l > 1 {
		l >>= 1
		h >>= 1
		for i := l; i <= h; i++ {
			m.tree[i] = HashInner(m.tree[i*2], m.tree[i*2+1])
		}
	}
}

// Build recomputes the entire mini merkle tree from the leaf-level slots.
// Cost: B-1 = 31 SHA256 calls for B=32.
func (m *MiniMerkle) Build() {
	m.buildRange(0, B-1)
}

// Clear sets all slots to the sentinel hash.
func (m *MiniMerkle) Clear() {
	for i := range m.tree {
		m.tree[i] = sentinelHash
	}
}

// SiblingPath returns the log₂(B) sibling hashes needed to prove that
// slot[index] is part of the mini merkle root. The path goes from the
// leaf level toward the root. Each entry is the sibling's hash at that level.
// Also returns the position indices (0=left child, 1=right child) indicating
// which side the proven slot is on at each level.
func (m *MiniMerkle) SiblingPath(index int) (siblings []Hash, positions []int) {
	siblings = make([]Hash, 0, miniMerkleDepth)
	positions = make([]int, 0, miniMerkleDepth)
	pos := B + index
	for pos > 1 {
		if pos%2 == 0 {
			// pos is left child, sibling is right
			siblings = append(siblings, m.tree[pos+1])
			positions = append(positions, 0) // proven node is left child
		} else {
			// pos is right child, sibling is left
			siblings = append(siblings, m.tree[pos-1])
			positions = append(positions, 1) // proven node is right child
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
