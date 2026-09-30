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
