// Package curchain_a is a test helper for multi-level cross-call chain tests.
// A is the middle hop: when called via cross, it cross-calls into curchain_b.
package curchain_a

import (
	"gno.land/r/tests/vm/curchain_b"
)

// A is a crossing function that cross-calls curchain_b.B(cross).
// Use from a filetest as: curchain_a.A(cross) to set up a 3-level chain
// caller → A → B with the filetest's main as the chain root.
func A(cur realm) {
	println("A cur.PkgPath:", cur.PkgPath())
	println("A prev.PkgPath:", cur.Previous().PkgPath())
	curchain_b.B(cross)
}
