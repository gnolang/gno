package main

// Nested array assignment must deep-copy (not shallow alias).

func main() {
	// Direct assignment
	a := [1][1]int{{1}}
	b := a
	b[0][0] = 2
	println(a[0][0], b[0][0])

	// Function argument passing
	c := [1][1]int{{1}}
	f(c)
	println(c[0][0])

	// Return by value
	d := [1][1]int{{1}}
	e := g(d)
	e[0][0] = 2
	println(d[0][0], e[0][0])
}

func f(x [1][1]int) {
	x[0][0] = 2
}

func g(a [1][1]int) [1][1]int {
	return a
}

// Output:
// 1 2
// 1
// 1 2
