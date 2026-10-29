package main

func main() {
	map[string]int{"a": 1, "b": 2} = map[string]int{"a": 1, "b": 2}
}

// Error:
// main/assign_literal5.gno:4:2-65: cannot assign to map[string]int{…}

// TypeCheckError:
// main/assign_literal5.gno:4:2: cannot assign to map[string]int{…} (neither addressable nor a map index expression)
