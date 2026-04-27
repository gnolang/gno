package main

func main() {
	a := 0
	if i := 1/a; true  {
		println(i)
	}
}

// Error:
// runtime error: division by zero

// Stacktrace:
// panic: runtime error: division by zero
// main<VPBlock(1,0)>()
//     main/panic5d.gno:5
