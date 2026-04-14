package main

func main() {
	a := 0
	if i := 1/a; true  {
		println(i)
	}
}

// Error:
// division by zero

// Stacktrace:
// panic: division by zero
// main<VPBlock(1,0)>()
//     main/panic5d.gno:5
