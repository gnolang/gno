package main

// A map's size argument is an advisory hint, and GnoVM ignores it
// entirely: the map is created empty and grows on insertion. Any hint —
// negative, or large enough to overflow internal size accounting — is a
// no-op that must never panic (unlike a bad slice length). Go also never
// panics on a bad map hint, though for valid hints it does preallocate.

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
