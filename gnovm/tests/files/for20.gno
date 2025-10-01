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

// TypeCheckError:
// main/for20.gno:6:1: label loop declared and not used
