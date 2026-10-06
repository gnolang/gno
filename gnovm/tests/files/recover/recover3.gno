package main

func main() {
	for i := 0; 1/i >= 0; i++ {
		break
	}
}

// Error:
// runtime error: division by zero

// Stacktrace:
// panic: runtime error: division by zero
// main<VPBlock(1,0)>()
//     main/recover3.gno:4
