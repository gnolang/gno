package main

func main() {
	a := 0
	if i := 1 / a; true {
		println(i)
	}
}

// Error:
// runtime error: integer divide by zero

// Stacktrace:
// panic: runtime error: integer divide by zero
// main<VPBlock(1,0)>()
//     main/panic5d.gno:5
