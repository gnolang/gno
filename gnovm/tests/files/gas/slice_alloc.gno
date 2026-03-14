// MAXALLOC: 500000000
package main

func main() {
	alloc(12499872) // 12499872 is the threshold to reach the allocation limit
}

func alloc(n int) {
	buffer := make([]int, n)
	buffer[0] = 1
}

// Gas:
// 500003087
