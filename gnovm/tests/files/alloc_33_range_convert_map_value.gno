// MAXALLOC: 1400000
package main

// Regression: a range's converted map value stays counted until written.
//
// `for k, v = range m` into a value target differing from the map's value
// type only by name copies the value to convert it, while the key needs no
// copy. The copy is a Go local until written, so only its anchor counts it
// when a write's own copy triggers a collection.
const N = 300000

type V [N]byte

func main() {
	m := map[int][N]byte{}
	var a [N]byte
	m[0] = a
	var k int
	var v V
	for k, v = range m {
	}
	println(k, len(v))
}

// Error:
// allocation limit exceeded
