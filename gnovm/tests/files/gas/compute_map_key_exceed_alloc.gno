// MAXALLOC: 500000000
package main

func main() {
	m := make(map[[500000001]byte]int)
	m[[500000001]byte{'0'}] = 1
}

// Error:
// allocation limit exceeded
