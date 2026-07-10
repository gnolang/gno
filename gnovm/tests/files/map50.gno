package main

type B struct {
	_, _ int
}

func main() {
	// Every value of B encodes to the same map key ({}), so all of
	// these refer to a single entry.
	m := map[B]string{B{1, 2}: "a"}
	println(m[B{3, 4}])

	m[B{5, 6}] = "b"
	println(len(m))
	println(m[B{0, 0}])

	delete(m, B{7, 8})
	println(len(m))
}

// Output:
// a
// 1
// b
// 0
