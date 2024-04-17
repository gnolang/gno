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
// main/files/types/0f14_filetest.gno:10: comparison on arrays of unequal type
