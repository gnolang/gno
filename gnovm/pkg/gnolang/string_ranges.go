package gnolang

// stringRangeSet is an ordered set of disjoint [start, end) string-backing
// extents keyed by start. It backs Allocator.stringRanges.
//
// It is a treap: insert, containment lookup and overlap check are
// O(log n) expected, so tracking cost per NewString stays logarithmic in
// the number of live tracked strings. A sorted slice was O(n) per insert
// (memmove), which made a tx that builds many small distinct strings
// quadratic in CPU while paying gas only for their bytes.
//
// Node priorities derive from the key (splitmix64 of start), so the
// tree shape is a pure function of the tracked set. Shape never affects
// accounting results, only CPU time.
//
// The zero value is an empty set.
type stringRangeSet struct {
	root *rangeNode
	n    int
}

type rangeNode struct {
	stringRange
	prio        uint64
	left, right *rangeNode
}

// rangePrio mixes start into a 64-bit priority (splitmix64 finalizer).
func rangePrio(start uintptr) uint64 {
	z := uint64(start) + 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (s *stringRangeSet) len() int { return s.n }

// insert adds r. The caller guarantees r is non-empty and does not
// overlap any tracked range (trackString clones on overlap first).
func (s *stringRangeSet) insert(r stringRange) {
	l, rt := rangeSplit(s.root, r.start)
	node := &rangeNode{stringRange: r, prio: rangePrio(r.start)}
	s.root = rangeMerge(rangeMerge(l, node), rt)
	s.n++
}

// containing returns the range with start <= p < end, or nil. The
// pointer is into the set; callers may update lastCycle through it.
func (s *stringRangeSet) containing(p uintptr) *stringRange {
	var best *rangeNode
	for n := s.root; n != nil; {
		if n.start <= p {
			best = n
			n = n.right
		} else {
			n = n.left
		}
	}
	if best == nil || p >= best.end {
		return nil
	}
	return &best.stringRange
}

// overlapping returns a range intersecting [p, end), or nil. Ranges are
// disjoint, so the one with the greatest start below end has the greatest
// end among candidates; if it does not reach p, none does.
func (s *stringRangeSet) overlapping(p, end uintptr) *stringRange {
	var best *rangeNode
	for n := s.root; n != nil; {
		if n.start < end {
			best = n
			n = n.right
		} else {
			n = n.left
		}
	}
	if best == nil || best.end <= p {
		return nil
	}
	return &best.stringRange
}

// retain keeps only the ranges with lastCycle == cycle and rebuilds the
// tree balanced. O(n); called once per GC cycle.
func (s *stringRangeSet) retain(cycle int64) {
	kept := make([]stringRange, 0, s.n)
	s.each(func(r *stringRange) {
		if r.lastCycle == cycle {
			kept = append(kept, *r)
		}
	})
	s.root = rangeBuild(kept, 0)
	s.n = len(kept)
}

// each visits the ranges in start order.
func (s *stringRangeSet) each(fn func(*stringRange)) {
	// Iterative in-order traversal; treap depth is O(log n) expected but
	// avoid recursion on the hot GC path regardless.
	var stack []*rangeNode
	for n := s.root; n != nil || len(stack) > 0; {
		for ; n != nil; n = n.left {
			stack = append(stack, n)
		}
		n = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		fn(&n.stringRange)
		n = n.right
	}
}

// ranges returns the tracked ranges in start order (tests).
func (s *stringRangeSet) ranges() []stringRange {
	out := make([]stringRange, 0, s.n)
	s.each(func(r *stringRange) { out = append(out, *r) })
	return out
}

// rangeSplit partitions t into (start < key, start >= key).
func rangeSplit(t *rangeNode, key uintptr) (l, r *rangeNode) {
	if t == nil {
		return nil, nil
	}
	if t.start < key {
		t.right, r = rangeSplit(t.right, key)
		return t, r
	}
	l, t.left = rangeSplit(t.left, key)
	return l, t
}

// rangeMerge joins l and r; every key in l precedes every key in r.
func rangeMerge(l, r *rangeNode) *rangeNode {
	if l == nil {
		return r
	}
	if r == nil {
		return l
	}
	if l.prio > r.prio {
		l.right = rangeMerge(l.right, r)
		return l
	}
	r.left = rangeMerge(l, r.left)
	return r
}

// rangeBuild makes a balanced tree from sorted ranges. Priorities decrease
// with depth so the heap invariant holds for later insert/remove.
func rangeBuild(sorted []stringRange, depth uint64) *rangeNode {
	if len(sorted) == 0 {
		return nil
	}
	mid := len(sorted) / 2
	return &rangeNode{
		stringRange: sorted[mid],
		prio:        ^uint64(0) - depth,
		left:        rangeBuild(sorted[:mid], depth+1),
		right:       rangeBuild(sorted[mid+1:], depth+1),
	}
}
