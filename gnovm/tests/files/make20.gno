package main

// An oversized fixed-size array is a compile-time error in Go ("larger than
// address space"). go/types does not catch it, so the gnovm preprocessor
// rejects it here instead of letting it leak to a runtime allocation panic.

func main() {
	var arr [9223372036854775807]int
	_ = arr
}

// Error:
// main/make20.gno:8:10-34: type [9223372036854775807]int larger than address space
