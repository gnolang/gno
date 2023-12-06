package main

// both not const, and both interface
func main() {
	println([]byte("a") == []byte("b"))
}

// Error:
