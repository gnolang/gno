// MAXALLOC: 500000000
package main

// Ranges over a [1000]int and a [1000]byte by value, key only and reading
// the element, through a pointer and as a slice. Only the range by value
// that reads an element copies the array, and that copy is charged once.
func main() {
	var a [1000]int
	for range a {
		break
	}
	for _, v := range a {
		_ = v
		break
	}
	for range &a {
		break
	}
	for range a[:] {
		break
	}
	var b [1000]byte
	for range b {
		break
	}
	for _, v := range b {
		_ = v
		break
	}
	for range &b {
		break
	}
	for range b[:] {
		break
	}
}

// Gas:
// 159584
