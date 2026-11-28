// MAXALLOC: 800000
package main

// Regression: a range's element stays counted while its key overwrites it.
//
// `for s[0], v = range s` reads s[0] before writing the key into it, so after
// the key write only the read element holds the array, and writing it to v
// copies it. The element is anchored until written, so the collection that
// copy triggers still counts it.
const N = 300000

func main() {
	var a [N]byte
	s := []any{a}
	var v any
	for s[0], v = range s {
	}
	println(len(v.([N]byte)))
}

// Error:
// allocation limit exceeded
