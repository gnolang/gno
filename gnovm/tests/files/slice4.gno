// MAXALLOC: 100000000
package main

func main() {
	b := []byte{1: 0x00, 1<<40 - 1: 0x00}
	_ = b // keep the variable used; the test is about something else
}

// Error:
// allocation limit exceeded
