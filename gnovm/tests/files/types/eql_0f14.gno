package main

var a [2]string
var c [2]int

// TODO: should stop at comparable check
func main() {
	a = [2]string{"hello", "world"}
	c = [2]int{1, 2}
	println(a == c)
}

// Error:
// main/eql_0f14.gno:10:10-16: invalid operation: (mismatched types [2]string and [2]int)

// TypeCheckError:
// main/eql_0f14.gno:10:15: invalid operation: a == c (mismatched types [2]string and [2]int)
