package main

type S struct {
	expected string
}

func main() {
	println("hello" == S{
		expected: `hello`[:],
	}.expected)
}

// Output:
// true
