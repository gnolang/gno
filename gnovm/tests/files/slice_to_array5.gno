package main

// Slice to array conversion - zero length array from any slice.

func main() {
	s := []byte{1, 2, 3}
	a := [0]byte(s)
	println(len(a))
}

// Output:
// 0
