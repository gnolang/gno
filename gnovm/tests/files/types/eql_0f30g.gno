package main

var a = [2]int{1, 1}
var b = []int{1, 1}

func main() {
	println(a == b)
}

// Error:
// main/files/types/eql_0f30g.gno:7: []int can only be compared to nil
