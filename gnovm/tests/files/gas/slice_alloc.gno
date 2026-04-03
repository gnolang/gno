// MAXALLOC: 500000000
package main

func main() {
	alloc(12499894) // 12499894 is the threshold to reach the allocation limit
}

func alloc(n int) {
	buffer := make([]int, n)
	buffer[0] = 1
}

// Gas:
// 7031

// Error:
// allocation limit exceeded
