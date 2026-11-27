// Package control imports bytes, so the reader's own package is trusted here.
package control

import (
	"bytes"
	"io"
)

var buf = make([]byte, 4)

// Fill reads a caller-supplied reader into this realm's buffer.
func Fill(cur realm, r io.ReaderAt) { r.ReadAt(buf, 0) }

// NewReader makes a reader over s, which keeps the bytes import in use.
func NewReader(s string) io.ReaderAt { return bytes.NewReader([]byte(s)) }

// State reports the buffer.
func State() string { return string(buf) }
