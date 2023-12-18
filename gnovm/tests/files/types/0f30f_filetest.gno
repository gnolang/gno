package main

var a = [2]int{1, 1}
var b = [3]int{1, 1}

func main() {
	println(a == b)
}

// Error:
// main/files/types/0f30f_filetest.gno:7: [2]int and [3]int cannot be compared
