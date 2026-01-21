package main

// Basic slice to array pointer conversion.

func main() {
	s := []byte{1, 2, 3, 4, 5}
	p := (*[5]byte)(s)
	println(p[0], p[2], p[4])
}

// Output:
// 1 3 5
