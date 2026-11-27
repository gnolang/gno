package main

func main() {
	b := []byte{1, 2, 3, 4, 5}
	println(b[6])
}

// Error:
// runtime error: index out of range [6] with length 5
