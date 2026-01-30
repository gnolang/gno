package main

var a = [2]int{1, 1}
var b = [3]int{1, 1}

func main() {
	println(a == b)
}

// Error:
// main/eql_0f30f.gno:7:10-16: invalid operation: (mismatched types [2]int and [3]int)

// TypeCheckError:
// main/eql_0f30f.gno:7:15: invalid operation: a == b (mismatched types [2]int and [3]int)
