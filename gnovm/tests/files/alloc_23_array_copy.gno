// MAXALLOC: 5000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// ArrayValue.Copy fills a fresh list one element copy at a time. Each element
// is a 300KB array, so a collection triggered by a later element's copy must
// still count the copies already in the list. The cap sits above what the
// program holds with the list left out of the count, and below what it holds
// with the list counted.
type Big [300000]byte

func main() {
	a := new([10]Big)
	b := *a
	println(len(b))
}

// Error:
// allocation limit exceeded
