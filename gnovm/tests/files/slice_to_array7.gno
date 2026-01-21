package main

// Slice to array conversion - nil slice to non-zero array should panic.

func main() {
	var s []byte
	a := [1]byte(s)
	_ = a
}

// Error:
// cannot convert slice with length 0 to array with length 1
