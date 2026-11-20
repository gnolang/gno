// Package viewgrant is a fixture for /p/ code that builds handles and
// closures: a /p/ method called by another realm cannot write the owner's
// stored object, though the owner's own code can.
package viewgrant

import "gno.land/p/nt/avl/v0"

// Writer is a handle a constructor hands out; its method writes only when the
// realm that built it calls it.
type Writer struct{ t *avl.Tree }

func NewWriter(t *avl.Tree) *Writer { return &Writer{t: t} }

func (w *Writer) Set(k, v string) { w.t.Set(k, v) }

// Deferred calls fn from a defer statement, for callback probes.
func (w *Writer) Deferred(fn func(string, any) bool) { defer fn("d", "deferred") }

// Wrap returns a closure built for the caller that calls fn.
func Wrap(fn func(string, any) bool) func() { return func() { fn("w", "wrapped") } }

// Pair holds two trees behind one heap-held struct; tests use it to show a
// /p/ method cannot write a tree another realm owns, while the owner's own
// crossing function can.
type Pair struct{ A, B *avl.Tree }

func NewPair() *Pair { return &Pair{A: avl.NewTree(), B: avl.NewTree()} }

func (p *Pair) SetA(k, v string) { p.A.Set(k, v) }
func (p *Pair) SetB(k, v string) { p.B.Set(k, v) }

// CopyTo writes its argument: library code calling it cannot write a tree
// owned by another realm.
func (p *Pair) CopyTo(dst *avl.Tree) { dst.Set("from", "pair") }
func (p *Pair) Get(k string) string {
	if v := p.A.Get(k); v != nil {
		return v.(string)
	}
	if v := p.B.Get(k); v != nil {
		return v.(string)
	}
	return ""
}
