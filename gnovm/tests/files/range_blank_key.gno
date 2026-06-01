package main

// Regression: `for _ = range slice` used to panic on nil key in assertIndexTypeIsInt.

func main() {
	s := []int{1, 2, 3}
	for _ = range s {
	}
}

// Output:
