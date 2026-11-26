// MAXALLOC: 500000000
package main

// A loop bounded by len of a [1000]int. len folds to a constant, so the
// condition makes no call on any iteration.
func main() {
	var a [1000]int
	for i := 0; i < len(a); i++ {
		a[i] = i
	}
}

// Gas:
// 1131329
