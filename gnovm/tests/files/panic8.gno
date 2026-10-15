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

// Error:
// runtime error: division by zero
