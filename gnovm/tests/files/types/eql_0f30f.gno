package main

var a = [2]int{1, 1}
var b = [3]int{1, 1}

func main() {
	println(a == b)
}

// Error:
// main/files/types/eql_0f30f.gno:7:10: cannot use [2]int as [3]int
