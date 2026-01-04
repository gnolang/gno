package main

func main() {
	a := 1
	b := 3
	println(a, b) // prints 1 3

	// Re-declaration of 'a' is allowed because 'c' is a new variable
	a, c := 2, 5
	println(a, c) // prints 2 5
}

// Output:
// 1 3
// 2 5
