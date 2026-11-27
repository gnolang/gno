// MAXALLOC: 1400000
package main

// Regression: a range's converted map key stays counted until written.
//
// `for k = range m` into a key target differing from the map's key type only
// by name copies the key to convert it, while no value is ranged. The copy is
// a Go local until written, so only its anchor counts it when the write's own
// copy triggers a collection.
const N = 300000

type K [N]byte

func main() {
	m := map[[N]byte]int{}
	var a [N]byte
	m[a] = 0
	var k K
	for k = range m {
	}
	println(len(k))
}

// Error:
// allocation limit exceeded
