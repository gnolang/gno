package main

type S struct {
	expected string
}

func main() {
	var s = S{
		expected: `hello`[:],
	}
	a := "hello"
	println(a == s.expected)
}

// Output:
// true
