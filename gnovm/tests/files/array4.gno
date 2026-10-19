package main

func main() {
	a := [3]int{2: 2, 1: 1, 2: 2}
	_ = a // keep the variable used; the test is about something else
}

// Error:
// main/array4.gno:4:26-27: duplicate index 2 in array or slice literal

// TypeCheckError:
// main/array4.gno:4:26: duplicate index 2 in array or slice literal
