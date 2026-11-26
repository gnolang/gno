// Package t72a imports t72b and bytes, and dispatches to a t72b value.
package t72a

import (
	"bytes"
	"io"

	"gno.land/p/demo/tests/t72b"
)

// Putter is the method set t72b.W satisfies.
type Putter interface{ Put([]byte) }

// NewB returns a t72b value behind an interface.
func NewB() Putter { return t72b.W{} }

// Apply makes the interface call from this package's code.
func Apply(p Putter, b []byte) { p.Put(b) }

// NewReader keeps the bytes import in use.
func NewReader(s string) io.ReaderAt { return bytes.NewReader([]byte(s)) }
