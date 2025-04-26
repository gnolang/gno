package main

func main() {
	map[string]int{"a": 1, "b": 2} = map[string]int{"a": 1, "b": 2}
}

// Error:
// main/files/types/assign_literal5.gno:4:2: cannot assign to map[(const-type string)] (const-type int){(const ("a" string)): (const (1 int)), (const ("b" string)): (const (2 int))}

// TypeCheckError:
// main/files/types/assign_literal5.gno:4:2: cannot assign to map[string]int{…} (neither addressable nor a map index expression)
