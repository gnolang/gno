package main

type f func() bool

var a f = func() bool { return true }
var b f = func() bool { return false }

func main() {
	println(a == b)
}

// Error:
// function can only be compared with `nil`
