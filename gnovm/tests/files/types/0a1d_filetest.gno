package main

type S struct {
	expected string
}

func main() {

	//a := "hello"

	println("hello" == S{
		expected: `hello`[:],
	}.expected)

}

// Output:
// true
