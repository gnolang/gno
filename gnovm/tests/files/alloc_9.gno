// MAXALLOC: 10000

package main

func main() {
	s := "hello"
	for i := 0; ; i++ {
		s += string(make([]byte, 150_000_000))
	}
}

// Error:
// allocation limit exceeded
