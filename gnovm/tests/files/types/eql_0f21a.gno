package main

type f func() bool

var a [2]*f
var b [2]*f

func main() {
	println(a == b)
}

// Output:
// true
