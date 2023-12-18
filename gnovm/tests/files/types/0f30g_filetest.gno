package main

var a = [2]int{1, 1}
var b = []int{1, 1}

func main() {
	println(a == b)
}

// Error:
// main/files/types/0f30g_filetest.gno:7: []int can only be compared to nil
