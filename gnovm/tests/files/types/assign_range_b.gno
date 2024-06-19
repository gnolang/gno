package main

import "fmt"

func main() {
	// Creating a map where keys are strings and values are integers
	scores := []string{
		"a", "b", "c",
	}

	// Using range to iterate over the map
	// k is the key and v is the value for each pair in the map
	var k, v int
	for k, v = range scores {
		fmt.Printf("%d has a score of %s\n", k, v)
	}
}

// Error:
// main/files/types/assign_range_b.gno:14: cannot use string as int
