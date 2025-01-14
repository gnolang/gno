package main

type S struct {
	expected string
}

// special case when RHS is result of slice operation, its type is determined in runtime
func main() {
	s := S{
		expected: `hello`[:], // this is not converted
	}

	a := "hello"

	println(a == s.expected)

}

// Output:
// true
