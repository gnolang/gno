package main

func main() {
	println("hello")
	defer func() {
		a := 0
		b := 1 / a
		_ = b // keep the variable used; the test is about something else
	}()
	println("world")
}

// Output:
// hello
// world

// Error:
// runtime error: division by zero

// Stacktrace:
// panic: runtime error: division by zero
// defer func(){ ... }()
//     main/panic8.gno:7
// main<VPBlock(1,0)>()
//     main/panic8.gno:5
