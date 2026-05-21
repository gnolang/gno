package main

func main() {
	// unkeyed elements followed by keyed element with large index
	a := []int{1, 2, 3, 11: 123}
	println(len(a), a[0], a[1], a[2], a[3], a[11])

	// keyed then unkeyed continuation
	b := []int{5: 50, 60, 70}
	println(len(b), b[0], b[5], b[6], b[7])

	// out-of-order keys
	c := []int{2: 20, 0: 5, 1: 10}
	println(len(c), c[0], c[1], c[2])

	// empty slice literal
	d := []int{}
	println(len(d))

	// string slice with gap
	e := []string{"a", "b", 5: "f"}
	println(len(e), e[0], e[1], e[2], e[5])
}

// Output:
// 12 1 2 3 0 123
// 8 0 50 60 70
// 3 5 10 20
// 0
// 6 a b  f
