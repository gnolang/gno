package main

func main() {
	println(map[string]int{"a": 1} == map[string]int{"b": 2})
}

// Error:
// main/files/types/eql_0f30a.gno:4:10: map[string]int can only be compared to nil

// TypeCheckError:
// main/files/types/eql_0f30a.gno:4:10: invalid operation: map[string]int{…} == map[string]int{…} (map can only be compared to nil)
