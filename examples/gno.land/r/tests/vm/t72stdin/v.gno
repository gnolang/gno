// Package t72stdin imports t72a, which imports bytes: the standard library
// is trusted anywhere in the realm's import tree.
package t72stdin

import (
	"io"

	"gno.land/p/demo/tests/t72a"
)

var buf = make([]byte, 4)

// Fill reads a caller-supplied reader into this realm's buffer.
func Fill(cur realm, r io.ReaderAt) { r.ReadAt(buf, 0) }

// Keep keeps the t72a import in use.
func Keep() t72a.Putter { return nil }

// State reports the buffer.
func State() string { return string(buf) }
