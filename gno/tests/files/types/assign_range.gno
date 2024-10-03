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
	for k, v := range scores {
		fmt.Printf("%s has a score of %d\n", k, v)
	}

	// Modifying values during iteration
	for k := range scores {
		scores[k] += 5 // Adding 5 to each score
	}
	fmt.Println("Updated scores:", scores)

	// Collecting keys and values into slices
	var keys []string
	var values []int
	for k, v := range scores {
		keys = append(keys, k)
		values = append(values, v)
	}
	fmt.Println("Keys:", keys)
	fmt.Println("Values:", values)
}

// Output:
// Alice has a score of 92
// Bob has a score of 89
// Charlie has a score of 95
// Updated scores: map[Alice:97 Bob:94 Charlie:100]
// Keys: [Alice Bob Charlie]
// Values: [97 94 100]
