package main

// Slice to array conversion - nil slice to zero-length array is OK.

func main() {
	var s []byte
	a := [0]byte(s)
	println(len(a))
}

// Output:
// 0
