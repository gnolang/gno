package main

type B struct {
	_, _ int
}

func main() {
	// All fields are blank, so no fields are compared and any two
	// values are equal under the Go spec.
	println(B{1, 2} == B{3, 4})
	println(B{1, 2} == B{1, 2})
}

// Output:
// true
// true
