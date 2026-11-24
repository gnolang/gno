// MAXALLOC: 4000000
package main

// Regression: a range's converted map key stops counting once written.
//
// `for k, v = range m` into targets differing from the map's types only by
// name converts both before writing either, and anchors both copies. Once the
// key is written, k holds it, so the collection the value write triggers must
// not count the key's copy a second time.
const NK = 300000

const NV = 600000

type K [NK]byte

type V [NV]byte

func main() {
	m := map[[NK]byte][NV]byte{}
	var a [NK]byte
	var b [NV]byte
	m[a] = b
	var k K
	var v V
	for k, v = range m {
	}
	println(len(k), len(v))
}

// Output:
// 300000 600000
