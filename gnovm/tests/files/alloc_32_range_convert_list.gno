// MAXALLOC: 1100000
package main

// Regression: a range's converted element stays counted until written.
//
// `for i, v = range s` into a target differing from the element's type only
// by name copies the element to convert it. The copy is a Go local until
// written, so only its anchor counts it when the write's own copy triggers a
// collection.
const N = 300000

type V [N]byte

func main() {
	s := [][N]byte{{}}
	var i int
	var v V
	for i, v = range s {
	}
	println(i, len(v))
}

// Error:
// allocation limit exceeded
