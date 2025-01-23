package main

func main() {
	// Ensure `break` works also when we have a label for the for loop.

loop:
	for {
		break
	}

	println("hey")
}

// Output:
// hey
