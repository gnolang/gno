package main

func main() {
	// unkeyed then keyed with gap
	a := []int{10, 20, 5: 50}
	println(len(a), a[0], a[1], a[2], a[5])

	// explicit key 0 followed by unkeyed
	b := []int{0: 100, 200, 300}
	println(len(b), b[0], b[1], b[2])

	// single keyed element with gap zero-fill
	c := []int{3: 42}
	println(len(c), c[0], c[1], c[2], c[3])
}

// Output:
// 6 10 20 0 50
// 3 100 200 300
// 4 0 0 0 42