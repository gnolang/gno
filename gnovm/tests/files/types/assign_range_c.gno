package main

import "fmt"

func main() {
	// Creating a map where keys are strings and values are integers
	scores := []string{
		"a", "b", "c",
	}

	// Using range to iterate over the map
	// k is the key and v is the value for each pair in the map
	var k, v float32
	for k, v = range scores {
		fmt.Printf("%d has a score of %s\n", k, v)
	}
}

// Error:
// main/assign_range_c.gno:14:2-16:3: index type should be int, but got float32

// TypeCheckError:
// main/assign_range_c.gno:14:6: cannot use k (value of type int) as float32 value in assignment; main/assign_range_c.gno:14:9: cannot use v (value of type string) as float32 value in assignment
