// Verifies that a promoted unexported field from another package is an
// access error, not a silently missing field.
package main

import "bytes"

type S struct {
	bytes.Buffer
}

func main() {
	var s S
	_ = s.buf
}

// Error:
// main/method44.gno:13:6-11: cannot access main.S.buf from main

// TypeCheckError:
// main/method44.gno:13:8: s.buf undefined (cannot refer to unexported field buf)
