package main

// Slice to array conversion - array smaller than slice.

func main() {
	s := []byte{1, 2, 3, 4, 5}
	a := [3]byte(s)
	println(a[0], a[1], a[2])
}

// Output:
// 1 2 3
