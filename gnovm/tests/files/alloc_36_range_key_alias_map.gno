// MAXALLOC: 800000
package main

// Regression: a range's map value stays counted while its key replaces it.
//
// `for m[0], v = range m` reads the entry's value before writing the key into
// m[0], so after the key write only the read value holds the array, and
// writing it to v copies it. The value is anchored until written, so the
// collection that copy triggers still counts it.
const N = 300000

func main() {
	m := map[int]any{}
	var a [N]byte
	m[0] = a
	var v any
	for m[0], v = range m {
	}
	println(len(v.([N]byte)))
}

// Error:
// allocation limit exceeded
