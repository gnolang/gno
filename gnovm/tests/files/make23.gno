package main

// An ellipsis-array key of MaxInt64 implies a length of MaxInt64+1, which
// overflows int64. Go rejects this key as an out-of-bounds array index (the
// largest key that forms a valid length is MaxInt64-1, used by make22), rather
// than as the "larger than address space" size error of make20/21/22. gno must
// reject it at compile time too, not wrap to a negative length.

func main() {
	x := [...]int{9223372036854775807: 1}
	_ = x
}

// Error:
// main/make23.gno:10:7-39: array index 9223372036854775807 out of bounds [0:0]
