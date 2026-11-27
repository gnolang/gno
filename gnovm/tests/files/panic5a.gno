package main

func main() {
	a := 0
	if i := 0; true && i/a == 1 {
		println("ok")
	}
}

// Stacktrace:
// panic: runtime error: integer divide by zero
// main<VPBlock(1,0)>()
//     main/panic5a.gno:5

// Error:
// runtime error: integer divide by zero
