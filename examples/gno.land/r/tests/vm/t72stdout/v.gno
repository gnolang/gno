// Package t72stdout imports only io: a caller-supplied bytes.Reader's package
// is not in this realm's import tree.
package t72stdout

import "io"

var buf = make([]byte, 4)

// Fill reads a caller-supplied reader into this realm's buffer.
func Fill(cur realm, r io.ReaderAt) { r.ReadAt(buf, 0) }

// State reports the buffer.
func State() string { return string(buf) }
