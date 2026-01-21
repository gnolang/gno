package main

// Basic slice to array conversion with exact length match.

func main() {
	s := []byte{1, 2, 3, 4, 5}
	a := [5]byte(s)
	println(a[0], a[2], a[4])
}

// Output:
// 1 3 5
