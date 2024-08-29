package main

import "math"

func main() {
	// Untyped constants are not addressable.
	_ = &math.MaxUint8
}

// Error:
// main/files/addressable_5a_err_stdlibs.gno:7:6: cannot take address of (const (255 <untyped> bigint))
