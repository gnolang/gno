// MAXALLOC: 500000000
package main

// Copies a [2000]int by value ten times. Each copy is charged per element.
//
// copy_array_1000.gno: same test with [1000]int
func copyOnce(src *[2000]int) int {
	c := *src
	return c[0]
}

func main() {
	var a [2000]int
	for i := 0; i < 10; i++ {
		_ = copyOnce(&a)
	}
}

// Gas:
// 1154822
