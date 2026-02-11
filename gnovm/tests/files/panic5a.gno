package main

func main() {
	a := 0
	if i := 0; true && i/a == 1 {
		println("ok")
	}
}

// Stacktrace:
// panic: division by zero
// main<VPBlock(1,0)>()
//     main/panic5a.gno:5

// Error:
// division by zero
