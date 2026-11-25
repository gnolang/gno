// MAXALLOC: 500000000
package main

// Copies a [1000]int by value ten times. Each copy is charged per element.
//
// copy_array_2000.gno: same test with [2000]int
func copyOnce(src *[1000]int) int {
	c := *src
	return c[0]
}

func main() {
	var a [1000]int
	for i := 0; i < 10; i++ {
		_ = copyOnce(&a)
	}
}

// Gas:
// 592922
