// MAXALLOC: 2300000
package main

// Regression: a range's converted key and value stay counted until written.
//
// `for k, v = range m` into targets differing from the map's types only by
// name copies the key and the value to convert them. The copies are Go locals
// until written, so only their anchor counts the key's copy when the value's
// copy, or a write's own copy, triggers a collection.
const N = 300000

type K [N]byte

type V [N]byte

func main() {
	m := map[[N]byte][N]byte{}
	var a [N]byte
	m[a] = a
	var k K
	var v V
	for k, v = range m {
	}
	println(len(k), len(v))
}

// Error:
// allocation limit exceeded
