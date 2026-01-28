package main

// Slice to array conversion - panic when slice is too short.

func main() {
	s := []byte{1, 2}
	a := [5]byte(s)
	_ = a
}

// Error:
// cannot convert slice with length 2 to array with length 5
