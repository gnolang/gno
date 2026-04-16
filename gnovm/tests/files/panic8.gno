package main

func main() {
	println("hello")
	defer func() {
		a := 0
		b := 1/a
	}()
	println("world")
}

// Output:
// hello
// world

// Error:
// runtime error: division by zero

// TypeCheckError:
// main/panic8.gno:7:3: declared and not used: b

// Stacktrace:
// panic: runtime error: division by zero
// defer func(){ ... }()
//     main/panic8.gno:7
// main<VPBlock(1,0)>()
//     main/panic8.gno:5
