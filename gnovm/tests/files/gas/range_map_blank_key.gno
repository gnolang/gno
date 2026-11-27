// MAXALLOC: 500000000
package main

// Ranges over a map whose three keys are [1000]int, with "_" for the key.
// Nothing reads "_", so no key is loaded or copied.

func main() {
	m := map[[1000]int]int{{1}: 1, {2}: 2, {3}: 3}
	n := 0
	for _, v := range m {
		n += v
	}
	println(n)
}

// Output:
// 6

// Gas:
// 554901
