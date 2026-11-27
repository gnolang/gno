package main

func main() {
	for i := 0; 1/i >= 0; i++ {
		break
	}
}

// Error:
// runtime error: integer divide by zero

// Stacktrace:
// panic: runtime error: integer divide by zero
// main<VPBlock(1,0)>()
//     main/recover3.gno:4
