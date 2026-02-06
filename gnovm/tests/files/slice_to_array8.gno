package main

// Slice to array conversion - conversion creates a copy (modifying original doesn't affect array).

func main() {
	s := []byte{1, 2, 3}
	a := [3]byte(s)
	s[0] = 99
	println(a[0], s[0])
}

// Output:
// 1 99
