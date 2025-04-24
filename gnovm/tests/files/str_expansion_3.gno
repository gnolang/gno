// MAXALLOC: 100000000

package main

func main() {
	s := "hello"
	s += string(make([]byte, 150_000_000))
}

// Error:
// allocation limit exceeded
