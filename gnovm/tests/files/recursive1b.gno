package main

type S struct {
	T *S
	B Integer
}

type Integer int

func main() {
	var a, b S
	println(a == b)
}

// Output:
// true
