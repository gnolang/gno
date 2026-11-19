// Package impvictim imports only io: a caller-supplied reader's own package is
// not in this realm's import tree, so its read runs view-only.
package impvictim

import "io"

var buf = make([]byte, 4)

// Fill reads a caller-supplied reader into this realm's buffer.
func Fill(cur realm, r io.ReaderAt) { r.ReadAt(buf, 0) }

// FillMV binds the reader's method value here, then calls it.
func FillMV(cur realm, r io.ReaderAt) {
	f := r.ReadAt
	f(buf, 0)
}

// FillFunc calls a method value the caller bound outside.
func FillFunc(cur realm, f func([]byte, int64) (int, error)) { f(buf, 0) }

// Filler hands out a crossing closure that writes this realm's buffer.
func Filler() func(realm, io.ReaderAt) {
	return func(cur realm, r io.ReaderAt) { r.ReadAt(buf, 0) }
}

// State reports the buffer.
func State() string { return string(buf) }
