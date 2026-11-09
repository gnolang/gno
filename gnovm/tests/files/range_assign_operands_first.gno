package main

// `for s[1], v = range s` reads each element before it writes the key, as an
// assignment evaluates its operands first: iteration 1 reads the 0 iteration 0
// wrote to s[1], before its own key writes 1 there. Go prints the same.
func main() {
	s := []int{10, 20, 30}
	var v int
	for s[1], v = range s {
		println(v)
	}
	// The map's one entry is read as 5 before its key, 0, is written to m[0].
	m := map[int]int{0: 5}
	for m[0], v = range m {
		println(v)
	}
}

// Output:
// 10
// 0
// 30
// 5
