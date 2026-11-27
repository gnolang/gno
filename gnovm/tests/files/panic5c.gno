package main

func main() {
	a := 0
	b := 1 / a
}

// Stacktrace:
// panic: runtime error: integer divide by zero
// main<VPBlock(1,0)>()
//     main/panic5c.gno:5

// Error:
// runtime error: integer divide by zero

// TypeCheckError:
// main/panic5c.gno:5:2: declared and not used: b
