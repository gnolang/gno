// Package viewgrant is a fixture for what /p/ code may do with mutable(x):
// grant an object it just built, never a stored one.
package viewgrant

import "gno.land/p/nt/avl/v0"

// Publish tries to grant the caller's stored object on its behalf.
func Publish(t *avl.Tree) *avl.Tree { return mutable(t) }

// Writer is a handle a constructor hands out, like a grc20 teller: its
// method writes the tree it wraps under the owner's authority.
type Writer struct{ t *avl.Tree }

func NewWriter(t *avl.Tree) *Writer { return mutable(&Writer{t: t}) }

func (w *Writer) Set(k, v string) { w.t.Set(k, v) }

// Deferred calls fn from a defer statement, for callback probes.
func (w *Writer) Deferred(fn func(string, any) bool) { defer fn("d", "deferred") }

// Wrap returns a closure built for the caller that calls fn.
func Wrap(fn func(string, any) bool) func() { return func() { fn("w", "wrapped") } }

// Pair holds two trees behind one heap-held struct, for grant-target probes.
type Pair struct{ A, B *avl.Tree }

func NewPair() *Pair { return &Pair{A: avl.NewTree(), B: avl.NewTree()} }

func (p *Pair) SetA(k, v string) { p.A.Set(k, v) }
func (p *Pair) SetB(k, v string) { p.B.Set(k, v) }
func (p *Pair) Get(k string) string {
	if v := p.A.Get(k); v != nil {
		return v.(string)
	}
	if v := p.B.Get(k); v != nil {
		return v.(string)
	}
	return ""
}
