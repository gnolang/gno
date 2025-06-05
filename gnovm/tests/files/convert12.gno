package main

func main() {
	s := []byte("hello")   // len(s) == 5
	arrPtr := ([5]byte)(s) // Convert to [5]byte
}

// Error:

// TypeCheckError:
