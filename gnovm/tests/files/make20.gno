package main

// A map hint large enough to overflow the allocator's internal size math
// (> ~1.15e17 = MaxInt64/allocMapItem) panics with a generic "size out of
// range" message rather than leaking the internal arithmetic-overflow error.

func main() {
	huge := int(^uint(0) >> 1) // max int
	_ = make(map[int]int, huge)
}

// Error:
// runtime error: makemap: size out of range
