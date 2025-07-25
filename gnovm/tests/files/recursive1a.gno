package main

type S1 *S

type S struct {
	T S1
}

func main() {
	var a, b S
	println(a == b)
}

// Output:
// true
