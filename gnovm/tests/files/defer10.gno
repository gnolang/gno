package main

func bar() (r int) {
	defer func() { // defer 1
		r += 4
		if recover() != nil {
			r += 3
		}
	}()

	var f func() // This is a nil function.
	// The nil function will be pushed onto the stack and executed later.
	// It will panic, and the panic will be caught by the defer statement above (defer 1).
	defer f()
	f = func() { // this has no effect since f (nil) is already pushed on the stack.
		r += 2
	}
	return 1
}

func main() {
	println(bar())
}

// Output:
// 8
