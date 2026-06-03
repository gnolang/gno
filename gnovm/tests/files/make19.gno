package main

// A map's size argument is an advisory hint. Mirroring Go's makemap, a
// negative or over-large hint must be silently treated as 0 (no
// preallocation), never panicking — including a hint big enough to overflow
// the allocator's internal size accounting.

func main() {
	neg := -1
	a := make(map[int]int, neg)
	println("neg:", len(a))

	huge := int(^uint(0) >> 1) // max int
	b := make(map[int]int, huge)
	println("huge:", len(b))

	a[1] = 10
	b[2] = 20
	println("usable:", a[1], b[2], len(a), len(b))
}

// Output:
// neg: 0
// huge: 0
// usable: 10 20 1 1
