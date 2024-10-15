package main

type f func() bool

var a f = func() bool { return true }

func main() {
	println(a == nil)
}

// Output:
// false
