// Package impvictim imports only io: a caller-supplied reader's own package is
// not in this realm's import tree, so its read runs view-only.
package impvictim

import "io"

var buf = make([]byte, 4)

// Fill reads a caller-supplied reader into this realm's buffer.
func Fill(cur realm, r io.ReaderAt) { r.ReadAt(buf, 0) }

// State reports the buffer.
func State() string { return string(buf) }
