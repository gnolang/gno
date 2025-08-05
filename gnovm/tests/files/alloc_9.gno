// MAXALLOC: 200000000
package main

import "fmt"

func grow() {
	var hugeString string
	for i := 0; i < 10; i++ {
		hugeString += string(make([]byte, 10_000_000))
	}
	totalMB := float64(len(hugeString)) / (1024 * 1024) // Convert to MB
	fmt.Printf("Total size: %.2f MB\n", totalMB)
}

func main() {
	grow()
}

// Output:
// Total size: 95.37 MB
