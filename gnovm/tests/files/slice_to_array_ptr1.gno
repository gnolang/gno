package main

// Slice to array pointer - modifying through pointer affects original slice.

func main() {
	s := []byte{1, 2, 3}
	p := (*[3]byte)(s)
	p[0] = 99
	println(s[0], p[0])
}

// Output:
// 99 99
