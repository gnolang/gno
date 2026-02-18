package main

func main() {
	a := 0
	b := 1/a
}

// Stacktrace:
// panic: division by zero
// main<VPBlock(1,0)>()
//     main/panic5c.gno:5

// Error:
// division by zero

// TypeCheckError:
// main/panic5c.gno:5:2: declared and not used: b
