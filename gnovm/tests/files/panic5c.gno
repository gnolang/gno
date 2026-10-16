package main

func main() {
	a := 0
	b := 1 / a
	_ = b // keep the variable used; the test is about something else
}

// Error:
// runtime error: division by zero
