package addrset

// ReadonlySet is a read-only view of a *Set. Cross-package callers cannot
// mutate the underlying set through this type: it exposes no mutator
// methods and holds the *Set in an unexported field, so a foreign realm
// can neither reach the set nor invoke Add/Remove on it.
//
// A ReadonlySet is a thin handle over the live Set (it does not copy or
// snapshot), so reads through it always reflect the Set's current contents.
type ReadonlySet struct {
	set *Set
}

// NewReadonlySet returns a read-only view of s.
func NewReadonlySet(s *Set) *ReadonlySet {
	return &ReadonlySet{set: s}
}

// Readonly returns a read-only view of the set.
func (s *Set) Readonly() *ReadonlySet {
	return NewReadonlySet(s)
}

// Has reports whether addr is in the underlying set.
func (r ReadonlySet) Has(addr address) bool {
	return r.set.Has(addr)
}

// Size returns the number of addresses in the underlying set.
func (r ReadonlySet) Size() int {
	return r.set.Size()
}

// Iterate walks the underlying set in sorted order, starting at offset and
// visiting up to count addresses. fn returns true to stop early; Iterate
// returns true if iteration was stopped that way.
//
// The wrapped Set.IterateByOffset has no return value, so the "stopped"
// result is synthesized from the last callback return via a
// closure-captured local.
func (r ReadonlySet) Iterate(offset, count int, fn func(addr address) bool) (stopped bool) {
	r.set.IterateByOffset(offset, count, func(a address) bool {
		stopped = fn(a)
		return stopped
	})
	return stopped
}
