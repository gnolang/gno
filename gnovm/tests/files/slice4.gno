// MAXALLOC: 100000000
package main

func main() {
	b := []byte{1: 0x00, 1<<40 - 1: 0x00}
}

// Error:
// allocation limit exceeded
