// MAXALLOC: 1300000
package main

// Regression: a range iteration that panics after anchoring its key and value,
// and is recovered, leaves nothing counted for the next range. The recovery
// drops its anchor, so the collection the next map range's key copy runs counts
// only what is live.
const B = 600000

const N = 300000

type K [1]byte

type K2 [N]byte

func trip() {
	defer func() { recover() }()
	m := map[[1]byte][B]byte{}
	m[[1]byte{}] = [B]byte{}
	var t []K
	var v [B]byte
	for t[0], v = range m {
	}
	_ = v
}

func garbage() {
	b := make([]byte, 500000)
	_ = b
}

func main() {
	trip()
	m2 := map[[N]byte]int{}
	m2[[N]byte{}] = 1
	var k K2
	var v int
	garbage()
	for k, v = range m2 {
	}
	println(len(k), v)
}

// Output:
// 300000 1
