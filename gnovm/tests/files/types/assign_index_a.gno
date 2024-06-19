package main

func main() {
	var s string
	var ok bool
	s, ok = map[string]int{"a": 1, "b": 2, "c": 3}["a"]
	println(s)
	println(ok)
}

// Error:
// main/files/types/assign_index_a.gno:6: cannot use int as string
