package main

var a = [2]int{1, 1}
var b = []int{1, 1}

func main() {
	println(a == b)
}

// Error:
// main/eql_0f30g.gno:7:10-16: []int can only be compared to nil

// TypeCheckError:
// main/eql_0f30g.gno:7:15: invalid operation: a == b (mismatched types [2]int and []int)
