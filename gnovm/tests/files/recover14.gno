package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	x, y := 10, 0
	_ = x / y // Panics because of division by zero
}

// Output:
// recover: division by zero
