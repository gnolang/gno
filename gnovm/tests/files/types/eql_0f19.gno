package main

type f func() bool

var a f

func main() {
	println(a == nil)
}

// Output:
// true
