package main

func main() {
	b := []byte{1, 2, 3, 4, 5}
	println(b[6])
}

// Error:
// slice index out of bounds: 6 (len=5)
