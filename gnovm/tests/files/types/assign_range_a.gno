package main

import "fmt"

func main() {
	// Creating a map where keys are strings and values are integers
	scores := map[string]int{
		"Alice":   92,
		"Bob":     89,
		"Charlie": 95,
	}

	// Using range to iterate over the map
	// k is the key and v is the value for each pair in the map
	var k, v int
	for k, v = range scores {
		fmt.Printf("%s has a score of %d\n", k, v)
	}
}

// Error:
// main/assign_range_a.gno:16:2-18:3: cannot use string as int

// TypeCheckError:
// main/assign_range_a.gno:16:6: cannot use k (value of type string) as int value in assignment
