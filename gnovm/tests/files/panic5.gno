package main

func main() {
	a := 0
	if i := 1; true && 
		i/a == 2 {
		println("ok")
	}
}

// Stacktrace:
// panic: runtime error: division by zero
// main<VPBlock(1,0)>()
//     main/panic5.gno:6

// Error:
// runtime error: division by zero
