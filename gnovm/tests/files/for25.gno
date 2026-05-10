package main

func main() {
	var s []int
	for ; s[0] == 0; {
		println("ok")
	}
}

// Stacktrace:
// panic: runtime error: nil slice index (out of bounds)
// main<VPBlock(1,0)>()
//     main/for25.gno:6

// Error:
// runtime error: nil slice index (out of bounds)
