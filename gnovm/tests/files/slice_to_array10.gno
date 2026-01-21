package main

// Slice to array conversion from sub-slice - tests offset handling.

func main() {
	s := make([]byte, 2, 4)
	s[0] = 10
	s[1] = 20
	a := [1]byte(s[1:])
	println(a[0] == s[1])
}

// Output:
// true
