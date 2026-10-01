// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

// doOpMapLit pops its operands and then copies each into a MapValue that is
// still a Go local. All ten values alias one 300KB package-level var, so the
// popped sources are a single object while the copies are ten fresh 300KB
// allocations. Without an anchor the fill trips the cap, GC recounts without
// seeing the map, the freed headroom is handed back, and 3MB is built under
// a 1MB cap. A MapValue's entries are a linked list rather than a
// []TypedValue, so this needs PushAnchorValue rather than PushAnchor.
func main() {
	m := map[int]Big{0: a, 1: a, 2: a, 3: a, 4: a, 5: a, 6: a, 7: a, 8: a, 9: a}
	println(len(m))
}

// Error:
// allocation limit exceeded
