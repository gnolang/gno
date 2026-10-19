package main

func main() {
	a := 0
	b := 1 / a
	_ = b // keep the variable used; the test is about something else
}

// Stacktrace:
// panic: runtime error: division by zero
// main<VPBlock(1,0)>()
//     main/panic5c.gno:5

// Error:
// runtime error: division by zero
