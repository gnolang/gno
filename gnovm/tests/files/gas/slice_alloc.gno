// MAXALLOC: 500000000
package main

func main() {
	alloc(12499838) // 12499838 is the largest n that stays under the allocation limit
}

func alloc(n int) {
	buffer := make([]int, n)
	buffer[0] = 1
}

// Gas:
// 570964336
