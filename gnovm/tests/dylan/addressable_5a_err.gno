package main

import "math"

func main() {
	// Untyped constants are not addressable.
	_ = &math.MaxUint8
}

// Error:
// main/dylan/addressable_5a_err.gno:7:6: cannot take address of (const (255 <untyped> bigint))
