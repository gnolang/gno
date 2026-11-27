package main

func main() {
	var s []int
	for s[0] == 0 {
		println("ok")
	}
}

// Stacktrace:
// panic: runtime error: index out of range [0] with length 0
// main<VPBlock(1,0)>()
//     main/for25.gno:6

// Error:
// runtime error: index out of range [0] with length 0
