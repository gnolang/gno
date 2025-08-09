package main

func main() {
	for i := 0; 1/i >= 0; i++ {
		break
	}
}

// Error:
// division by zero

// Stacktrace:
// panic: division by zero
// main<VPBlock(1,0)>()
//     main/recover3.gno:4
