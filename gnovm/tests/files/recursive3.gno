package main

type S struct {
	T *S
}

func main() {
	var a, b S
	println(a == b)
}

// Output:
// true
