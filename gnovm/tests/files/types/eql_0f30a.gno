package main

func main() {
	println(map[string]int{"a": 1} == map[string]int{"b": 2})
}

// Error:
// main/files/types/eql_0f30a.gno:4: map[string]int can only be compared to nil
