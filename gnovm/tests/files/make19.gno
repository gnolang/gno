package main

// A negative map hint is silently treated as 0 (no preallocation), mirroring
// Go's makemap. The map remains fully usable afterwards.

func main() {
	neg := -1
	a := make(map[int]int, neg)
	println("neg:", len(a))

	a[1] = 10
	println("usable:", a[1], len(a))
}

// Output:
// neg: 0
// usable: 10 1
