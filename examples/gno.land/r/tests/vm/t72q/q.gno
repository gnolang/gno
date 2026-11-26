// Package t72q imports t72b and hands out its value; a realm importing t72q
// does not import t72b itself.
package t72q

import "gno.land/p/demo/tests/t72b"

// NewB returns a t72b value behind an interface.
func NewB() interface{ Put([]byte) } { return t72b.W{} }
