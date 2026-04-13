package main

var a = 0
func foo()int{
	return 1/a
}
func main() {
	foo()
}

// Error:
// division by zero

// Stacktrace:
// panic: division by zero
// foo<VPBlock(3,1)>()
//     main/panic3b.gno:5
// main<VPBlock(1,2)>()
//     main/panic3b.gno:8
