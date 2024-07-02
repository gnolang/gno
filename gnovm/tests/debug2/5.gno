package main

import "fmt"

func main() {
	var counter int
	var fns []func()

	defer func() {
		for _, f := range fns {
			f()
		}
	}()

LABEL_1:
	if counter >= 2 {
		goto LABEL_2 // Jump to LABEL_2 once counter is 2 or more
	}

	// capture
	fns = append(fns, func() { println(counter) })
	counter++
	goto LABEL_1 // Go back to the beginning of the loop

LABEL_2:
	fmt.Println("Exited loop with counter at:", counter)
	// Continue with the rest of the program
}

// Output:
// Exited loop with counter at: 2
// 1
// 2
