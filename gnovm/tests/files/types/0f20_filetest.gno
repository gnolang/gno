package main

type f func() bool

var a [2]f
var b [2]f

// TODO: review where it happens
func main() {
	println(a == nil)
}

// Error:
// interface conversion: gnolang.Value is nil, not *gnolang.ArrayValue
